import { Member } from "./member";

export interface TeamspaceResponse {
    Teamspace: Teamspace;
}

export interface Teamspace {
    ID: string;
    Name: string;
    Description?: string;
    CreatedAt?: string;
    CreatorID: string;
    Members?: Member[] | null;
}