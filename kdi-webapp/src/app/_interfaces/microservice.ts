import { Container } from "./container";

export interface Microservice {
    ID?: string;
    Name?: string;
    namespace: string;
    Replicas?: Int16Array;
    Labels?: string[];
    Selectors?: string[];
    Strategy?: string;
    CreatorID?: string;
    DeployedAt?: Date;
    EnvironmentID?: string;
    Containers: Container[];
    Conditions: Conditions[]
}

export interface Conditions {
    type?: string;
    message?: string;
    reason?: string;
}

