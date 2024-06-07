import { Container } from "./container";

export interface Microservice {
    ID?: string;
    Name: string;
    NamespaceID?: string;
    Replicas: Int16Array;
    Labels?: string[];
    Selectors?: string[];
    Strategy: string;
    Status: Date;
    CreatorID?: string;
    EnvironmentID?: string;
    Containers?: Container[];
    Conditions?: Conditions[]
}

export interface Conditions {
    Type?: string;
    Status: string;
    Reason?: string;
}

