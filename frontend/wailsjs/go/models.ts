export namespace fetcher {
	
	export class QuotaResult {
	    platform: string;
	    id: string;
	    abbr: string;
	    kind: string;
	    used: number;
	    total: number;
	    percent: number;
	    balance: number;
	    currency: string;
	    remaining: string;
	    reset_at: string;
	    // Go type: time
	    last_updated: any;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new QuotaResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.id = source["id"];
	        this.abbr = source["abbr"];
	        this.kind = source["kind"];
	        this.used = source["used"];
	        this.total = source["total"];
	        this.percent = source["percent"];
	        this.balance = source["balance"];
	        this.currency = source["currency"];
	        this.remaining = source["remaining"];
	        this.reset_at = source["reset_at"];
	        this.last_updated = this.convertValues(source["last_updated"], null);
	        this.error = source["error"];
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

export namespace main {
	
	export class ProviderInput {
	    id: string;
	    enabled: boolean;
	    creds: Record<string, string>;
	    budget: number;
	
	    static createFrom(source: any = {}) {
	        return new ProviderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.creds = source["creds"];
	        this.budget = source["budget"];
	    }
	}

}

