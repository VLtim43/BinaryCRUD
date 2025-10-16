export namespace main {
	
	export class ItemDTO {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class OrderItemDTO {
	    itemId: number;
	    quantity: number;
	
	    static createFrom(source: any = {}) {
	        return new OrderItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.quantity = source["quantity"];
	    }
	}

}

