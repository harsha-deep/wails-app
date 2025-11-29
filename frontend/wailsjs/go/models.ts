export namespace main {
	
	export class CPUStats {
	    usage: number;
	    cores: number;
	    modelName: string;
	    user: number;
	    system: number;
	    idle: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usage = source["usage"];
	        this.cores = source["cores"];
	        this.modelName = source["modelName"];
	        this.user = source["user"];
	        this.system = source["system"];
	        this.idle = source["idle"];
	    }
	}
	export class MemoryStats {
	    total: number;
	    available: number;
	    used: number;
	    usedPercent: number;
	    free: number;
	    buffers: number;
	    cached: number;
	    swapTotal: number;
	    swapFree: number;
	    swapUsed: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.available = source["available"];
	        this.used = source["used"];
	        this.usedPercent = source["usedPercent"];
	        this.free = source["free"];
	        this.buffers = source["buffers"];
	        this.cached = source["cached"];
	        this.swapTotal = source["swapTotal"];
	        this.swapFree = source["swapFree"];
	        this.swapUsed = source["swapUsed"];
	    }
	}
	export class ProcessInfo {
	    pid: number;
	    name: string;
	    state: string;
	    cpu: number;
	    memory: number;
	    threads: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.state = source["state"];
	        this.cpu = source["cpu"];
	        this.memory = source["memory"];
	        this.threads = source["threads"];
	    }
	}
	export class SystemStats {
	    cpu: CPUStats;
	    memory: MemoryStats;
	    processes: ProcessInfo[];
	    uptime: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], CPUStats);
	        this.memory = this.convertValues(source["memory"], MemoryStats);
	        this.processes = this.convertValues(source["processes"], ProcessInfo);
	        this.uptime = source["uptime"];
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

}

