import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import { AppProxy } from "./constructs/app-proxy";

export interface AppProxyStackProps extends cdk.StackProps {
    readonly tenant: string;
    readonly environment: string;
    readonly product: string;
    readonly workspaceUri: string;
    readonly appUri: string;
}

export class AppProxyStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props: AppProxyStackProps) {
        super(scope, id, props);

        new AppProxy(this, "AppProxy", {
            tenant: props.tenant,
            environment: props.environment,
            product: props.product,
            workspaceUri: props.workspaceUri,
            appUri: props.appUri,
        });
    }
}
