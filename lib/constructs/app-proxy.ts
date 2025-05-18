import { Construct } from "constructs";
import { CfnOutput, CfnParameter, Duration, RemovalPolicy, SecretValue } from "aws-cdk-lib";
import {
    ContainerImage,
    Cluster,
    OperatingSystemFamily,
    CpuArchitecture,
    CfnService,
    FargateTaskDefinition,
    LinuxParameters,
    LogDrivers,
    Protocol,
    Secret as EcsSecret,
} from "aws-cdk-lib/aws-ecs";
import { StringParameter } from "aws-cdk-lib/aws-ssm";
import { Vpc, SecurityGroup } from "aws-cdk-lib/aws-ec2";
import { PolicyStatement, Effect } from "aws-cdk-lib/aws-iam";
import {
    ApplicationLoadBalancedFargateService,
    ApplicationLoadBalancedServiceRecordType,
} from "aws-cdk-lib/aws-ecs-patterns";
import { LogGroup, RetentionDays } from "aws-cdk-lib/aws-logs";
import { Secret } from "aws-cdk-lib/aws-secretsmanager";
import { HostedZone, ARecord, AaaaRecord, RecordTarget } from "aws-cdk-lib/aws-route53";
import { Certificate, CertificateValidation } from "aws-cdk-lib/aws-certificatemanager";

import config from "../config/constants";
import { LoadBalancerTarget } from "aws-cdk-lib/aws-route53-targets";
import { ApplicationProtocol, SslPolicy } from "aws-cdk-lib/aws-elasticloadbalancingv2";

export interface AppProxyProps {
    readonly tenant: string;
    readonly environment: string;
    readonly product: string;
    readonly workspaceUri: string;
    readonly appUri: string;
    readonly hostname: string;
    readonly zone: string;
}

export class AppProxy extends Construct {
    constructor(scope: Construct, id: string, props: AppProxyProps) {
        super(scope, id);

        const clientID = new CfnParameter(this, "clientID", {
            type: "String",
            noEcho: true,
            default: process.env.SERVICE_PRINCIPLE_CLIENT_ID ?? "unset",
        });

        const clientSecret = new CfnParameter(this, "clientSecret", {
            type: "String",
            noEcho: true,
            default: process.env.SERVICE_PRINCIPLE_CLIENT_SECRET ?? "unset",
        });

        const servicePrincipleCreds = new Secret(this, "CredsSecret", {
            secretObjectValue: {
                client_id: SecretValue.cfnParameter(clientID),
                client_secret: SecretValue.cfnParameter(clientSecret),
            },
            description: "Databricks Service Principle Creds",
            removalPolicy: RemovalPolicy.DESTROY,
        });

        const vpc = Vpc.fromLookup(this, "ImportVPC", {
            isDefault: false,
            vpcId: StringParameter.valueFromLookup(scope, `/${props.tenant}/baseline/network/vpc-id`),
        });

        const cluster = new Cluster(this, "Cluster", { vpc: vpc });

        // setup our task def basics
        const taskDefinition = new FargateTaskDefinition(this, "TaskDef", {
            cpu: config.task.cpu,
            memoryLimitMiB: config.task.memory,
            runtimePlatform: {
                operatingSystemFamily: OperatingSystemFamily.LINUX,
                cpuArchitecture: CpuArchitecture.ARM64,
            },
        });

        taskDefinition.addToExecutionRolePolicy(
            new PolicyStatement({
                effect: Effect.ALLOW,
                resources: ["*"],
                actions: [
                    "ecr:GetAuthorizationToken",
                    "ecr:BatchCheckLayerAvailability",
                    "ecr:GetDownloadUrlForLayer",
                    "ecr:BatchGetImage",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents",
                ],
            }),
        );

        taskDefinition.addToExecutionRolePolicy(
            new PolicyStatement({
                effect: Effect.ALLOW,
                resources: [servicePrincipleCreds.secretArn],
                actions: ["secretsmanager:GetSecretValue"],
            }),
        );

        // ssm permissions
        taskDefinition.addToTaskRolePolicy(
            new PolicyStatement({
                resources: ["*"],
                actions: [
                    "ssmmessages:CreateControlChannel",
                    "ssmmessages:CreateDataChannel",
                    "ssmmessages:OpenControlChannel",
                    "ssmmessages:OpenDataChannel",
                ],
                effect: Effect.ALLOW,
            }),
        );

        const logGroup = new LogGroup(this, "LogGroup", {
            logGroupName: `/${props.tenant.toLowerCase()}/${props.product.toLowerCase()}/${props.environment.toLowerCase()}/ecs`,
            retention: RetentionDays.ONE_WEEK,
            removalPolicy: RemovalPolicy.DESTROY,
        });

        const container = taskDefinition.addContainer("Caddy", {
            image: ContainerImage.fromAsset("./resources/app-proxy", {
                buildArgs: {
                    CADDY_VERSION: config.versions.CADDY,
                    GO_VERSION: config.versions.GO,
                    BUILD_DATE: process.env.DATE ?? "19700101",
                    VCS_REF: process.env.COMMIT ?? "aaaaaaaa",
                },
            }),
            logging: LogDrivers.awsLogs({ streamPrefix: "DbxAppProxy", logGroup: logGroup }),
            containerName: "appproxy",
            linuxParameters: new LinuxParameters(this, "LinuxParams", {
                initProcessEnabled: true,
            }),
            healthCheck: {
                command: ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
                interval: Duration.seconds(60),
                retries: 3,
                startPeriod: Duration.seconds(60),
                timeout: Duration.seconds(5),
            },
            environment: {
                LOG_LEVEL: "INFO",
                DATABRICKS_WORKSPACE_URI: props.workspaceUri,
                DATABRICKS_APP_URI: props.appUri,
            },
            secrets: {
                DATABRICKS_CLIENT_ID: EcsSecret.fromSecretsManager(servicePrincipleCreds, "client_id"),
                DATABRICKS_CLIENT_SECRET: EcsSecret.fromSecretsManager(servicePrincipleCreds, "client_secret"),
            },
        });

        // API port
        container.addPortMappings({
            containerPort: 8080,
            hostPort: 8080,
            protocol: Protocol.TCP,
        });

        // our container security group
        const serviceSG = new SecurityGroup(this, "SecurityGroup", {
            vpc: vpc,
            allowAllOutbound: true,
            allowAllIpv6Outbound: true,
            description: "Security group for the service",
        });

        serviceSG.node.addDependency(vpc);

        const zone = HostedZone.fromLookup(this, "Zone", {
            domainName: props.zone,
        });

        const proxyName = `${props.hostname}.${zone.zoneName}`;

        const certificate = new Certificate(this, "SiteCertificate", {
            domainName: proxyName,
            validation: CertificateValidation.fromDns(zone),
        });

        // our service cluster
        const loadBalancedFargateService = new ApplicationLoadBalancedFargateService(this, "Service", {
            cluster: cluster,
            circuitBreaker: {
                rollback: true,
            },
            desiredCount: config.task.count,
            publicLoadBalancer: true,
            securityGroups: [serviceSG],
            capacityProviderStrategies: [
                {
                    capacityProvider: "FARGATE",
                    weight: 1,
                },
            ],
            protocol: ApplicationProtocol.HTTPS,
            domainName: proxyName,
            domainZone: zone,
            recordType: ApplicationLoadBalancedServiceRecordType.NONE,
            certificate: certificate,
            redirectHTTP: true,
            taskDefinition: taskDefinition,
            taskSubnets: {
                subnets: vpc.privateSubnets,
            },
            loadBalancerName: "AppProxyLB",
            sslPolicy: SslPolicy.TLS13_RES,
        });

        // speed up cluster deploys
        loadBalancedFargateService.targetGroup.setAttribute("deregistration_delay.timeout_seconds", "10");

        loadBalancedFargateService.targetGroup.configureHealthCheck({
            path: "/health",
            interval: Duration.seconds(30),
            healthyThresholdCount: 3,
            unhealthyThresholdCount: 6,
        });

        // allow ecs exec to cluster
        const cfnService = loadBalancedFargateService.service.node.defaultChild as CfnService;
        cfnService.enableExecuteCommand = true;

        new ARecord(this, "ProxyAliasRecordA", {
            zone: zone,
            recordName: proxyName,
            target: RecordTarget.fromAlias(new LoadBalancerTarget(loadBalancedFargateService.loadBalancer, {})),
        });

        new AaaaRecord(this, "ProxyAliasRecordAAAA", {
            zone: zone,
            recordName: proxyName,
            target: RecordTarget.fromAlias(new LoadBalancerTarget(loadBalancedFargateService.loadBalancer, {})),
        });

        new CfnOutput(this, "ProxyURI", {
            value: `${loadBalancedFargateService.loadBalancer.loadBalancerDnsName}`,
            description: "The App Proxy Endpoint",
        });
    }
}
