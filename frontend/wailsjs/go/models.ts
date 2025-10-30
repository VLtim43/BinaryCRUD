export namespace dao {
	
	export class OrderDTO {
	    id: number;
	    items: string[];
	
	    static createFrom(source: any = {}) {
	        return new OrderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.items = source["items"];
	    }
	}
	export class PromotionDTO {
	    id: number;
	    name: string;
	    items: string[];
	
	    static createFrom(source: any = {}) {
	        return new PromotionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.items = source["items"];
	    }
	}

}

export namespace main {
	
	export class ItemDTO {
	    id: number;
	    name: string;
	    priceInCents: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.priceInCents = source["priceInCents"];
	    }
	}
	export class LogEntry {
	    timestamp: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.message = source["message"];
	    }
	}

}

