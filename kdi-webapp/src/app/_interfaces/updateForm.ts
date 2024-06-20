export interface UpdateForm{
    strategy?: string,
    replicas: Number,
    image: string,
    maxUnavailable: string,
    maxSurge: string,
    canaryWeight: string,
    canaryAnalysisInterval:string,
}