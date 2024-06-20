export interface Project {
    ID: string;
    Name: string;
    Description: string;
    CreatedAt: Date;
    CreatorID: string;
    TeamspaceID: string;
    Owner?: string;
}