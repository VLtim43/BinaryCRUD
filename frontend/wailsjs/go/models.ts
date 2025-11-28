export namespace main {
	
	export class CompactResult {
	    itemsRemoved: number;
	    ordersAffected: number;
	    promotionsAffected: number;
	    ordersRemoved: number;
	    promotionsRemoved: number;
	    orderPromotionsRemoved: number;
	
	    static createFrom(source: any = {}) {
	        return new CompactResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemsRemoved = source["itemsRemoved"];
	        this.ordersAffected = source["ordersAffected"];
	        this.promotionsAffected = source["promotionsAffected"];
	        this.ordersRemoved = source["ordersRemoved"];
	        this.promotionsRemoved = source["promotionsRemoved"];
	        this.orderPromotionsRemoved = source["orderPromotionsRemoved"];
	    }
	}
	export class LogEntry {
	    timestamp: string;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.message = source["message"];
	    }
	}

}

