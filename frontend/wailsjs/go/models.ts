export namespace binding {
	
	export class ActivateRequestDTO {
	    skillId: string;
	    agent: string;
	    scope: string;
	    projectId: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivateRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.agent = source["agent"];
	        this.scope = source["scope"];
	        this.projectId = source["projectId"];
	    }
	}
	export class ConflictDTO {
	    skillId: string;
	    agent: string;
	    globalActivation?: ActivationDTO;
	    projectActivation?: ActivationDTO;
	
	    static createFrom(source: any = {}) {
	        return new ConflictDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.agent = source["agent"];
	        this.globalActivation = this.convertValues(source["globalActivation"], ActivationDTO);
	        this.projectActivation = this.convertValues(source["projectActivation"], ActivationDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ActivationDTO {
	    id: number;
	    skillId: string;
	    agent: string;
	    scope: string;
	    projectId: string;
	    appliedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.skillId = source["skillId"];
	        this.agent = source["agent"];
	        this.scope = source["scope"];
	        this.projectId = source["projectId"];
	        this.appliedAt = source["appliedAt"];
	    }
	}
	export class ActivateResultDTO {
	    activation?: ActivationDTO;
	    conflict?: ConflictDTO;
	
	    static createFrom(source: any = {}) {
	        return new ActivateResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activation = this.convertValues(source["activation"], ActivationDTO);
	        this.conflict = this.convertValues(source["conflict"], ConflictDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ActivationFilterDTO {
	    skillId: string;
	    agent: string;
	    scope: string;
	    projectId: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivationFilterDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.agent = source["agent"];
	        this.scope = source["scope"];
	        this.projectId = source["projectId"];
	    }
	}
	export class SkillProjectRef {
	    id: string;
	    name: string;
	    path: string;
	    skillPath: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillProjectRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.skillPath = source["skillPath"];
	    }
	}
	export class AggregatedSkillDTO {
	    name: string;
	    description: string;
	    categoryId?: number;
	    categoryName: string;
	    isGlobal: boolean;
	    globalPath: string;
	    projects: SkillProjectRef[];
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AggregatedSkillDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.categoryId = source["categoryId"];
	        this.categoryName = source["categoryName"];
	        this.isGlobal = source["isGlobal"];
	        this.globalPath = source["globalPath"];
	        this.projects = this.convertValues(source["projects"], SkillProjectRef);
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AssignSkillCategoryRequestDTO {
	    skillName: string;
	    skillPath: string;
	    categoryId?: number;
	
	    static createFrom(source: any = {}) {
	        return new AssignSkillCategoryRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.skillPath = source["skillPath"];
	        this.categoryId = source["categoryId"];
	    }
	}
	export class CategoryDTO {
	    id: number;
	    name: string;
	    description: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CategoryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.createdAt = source["createdAt"];
	    }
	}
	
	export class CopySkillRequestDTO {
	    skillId: string;
	    sourceProjectId: string;
	    targetProjectId: string;
	    agent: string;
	
	    static createFrom(source: any = {}) {
	        return new CopySkillRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.sourceProjectId = source["sourceProjectId"];
	        this.targetProjectId = source["targetProjectId"];
	        this.agent = source["agent"];
	    }
	}
	export class CreateCategoryRequestDTO {
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateCategoryRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class DeleteSkillRequestDTO {
	    skillId: string;
	    projectId: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteSkillRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.projectId = source["projectId"];
	    }
	}
	export class DoctorIssueDTO {
	    kind: string;
	    title: string;
	    detail: string;
	    howToFix: string;
	    fixable: boolean;
	    fixData: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new DoctorIssueDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.howToFix = source["howToFix"];
	        this.fixable = source["fixable"];
	        this.fixData = source["fixData"];
	    }
	}
	export class DoctorReportDTO {
	    issues: DoctorIssueDTO[];
	
	    static createFrom(source: any = {}) {
	        return new DoctorReportDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.issues = this.convertValues(source["issues"], DoctorIssueDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectCandidateDTO {
	    name: string;
	    path: string;
	    detectedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectCandidateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.detectedAgents = source["detectedAgents"];
	    }
	}
	export class ProjectCategoryLinkDTO {
	    projectId: string;
	    categoryId: number;
	    agent: string;
	    category: CategoryDTO;
	
	    static createFrom(source: any = {}) {
	        return new ProjectCategoryLinkDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.categoryId = source["categoryId"];
	        this.agent = source["agent"];
	        this.category = this.convertValues(source["category"], CategoryDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectCategoryRequestDTO {
	    projectId: string;
	    categoryId: number;
	    agent: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectCategoryRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.categoryId = source["categoryId"];
	        this.agent = source["agent"];
	    }
	}
	export class ProjectDTO {
	    id: string;
	    name: string;
	    path: string;
	    detectedAgents: string[];
	    addedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.detectedAgents = source["detectedAgents"];
	        this.addedAt = source["addedAt"];
	    }
	}
	export class RegisterProjectRequestDTO {
	    path: string;
	    detectedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new RegisterProjectRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.detectedAgents = source["detectedAgents"];
	    }
	}
	export class ResolveConflictRequestDTO {
	    conflict: ConflictDTO;
	    resolution: number;
	
	    static createFrom(source: any = {}) {
	        return new ResolveConflictRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conflict = this.convertValues(source["conflict"], ConflictDTO);
	        this.resolution = source["resolution"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SkillDTO {
	    id: string;
	    name: string;
	    description: string;
	    categoryId?: number;
	    categoryName: string;
	    path: string;
	    source: string;
	    ownerProjectId: string;
	    ownerProjectName: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.categoryId = source["categoryId"];
	        this.categoryName = source["categoryName"];
	        this.path = source["path"];
	        this.source = source["source"];
	        this.ownerProjectId = source["ownerProjectId"];
	        this.ownerProjectName = source["ownerProjectName"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	
	export class UpdateCategoryRequestDTO {
	    id: number;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCategoryRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}

}

export namespace config {
	
	export class Settings {
	    workspaceRoots: string[];
	    globalSkillSources: string[];
	    skillsHome?: string;
	    skillSources?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceRoots = source["workspaceRoots"];
	        this.globalSkillSources = source["globalSkillSources"];
	        this.skillsHome = source["skillsHome"];
	        this.skillSources = source["skillSources"];
	    }
	}

}

