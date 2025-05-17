#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { AppProxyStack } from "../lib/app-proxy-stack";
import { capitalise } from "../lib/utils";

const app = new cdk.App();

// infra namespacing
const tenant = process.env.TENANT ?? "Abc";
const product = process.env.PRODUCT ?? "DbxAppProxy";
const environment = process.env.ENVIRONMENT ?? "Dev";

// app configuration
const workspaceUri = process.env.WORKSPACE_URI ?? "https://localhost";
const appUri = process.env.APP_URI ?? "https://localhost";

const stack = new AppProxyStack(app, capitalise(tenant) + capitalise(product) + capitalise(environment), {
    tenant: tenant,
    environment: environment,
    product: product,
    workspaceUri: workspaceUri,
    appUri: appUri,
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION ?? "eu-west-1" },
});

cdk.Tags.of(stack).add("tenant", tenant, {});
cdk.Tags.of(stack).add("environment", environment, {});
cdk.Tags.of(stack).add("product", product, {});
