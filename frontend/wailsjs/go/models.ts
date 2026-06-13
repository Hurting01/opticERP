export namespace models {
	
	export class Position {
	    id: number;
	    name: string;
	    created_at: string;
	    norm_hours?: number;
	    hours_per_shift?: number;
	    salary?: number;
	    additional_payments?: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = source["created_at"];
	        this.norm_hours = source["norm_hours"];
	        this.hours_per_shift = source["hours_per_shift"];
	        this.salary = source["salary"];
	        this.additional_payments = source["additional_payments"];
	    }
	}
	export class StaffWithPosition {
	    id: number;
	    full_name: string;
	    position_id: number;
	    position_name: string;
	    is_active: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new StaffWithPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.full_name = source["full_name"];
	        this.position_id = source["position_id"];
	        this.position_name = source["position_name"];
	        this.is_active = source["is_active"];
	        this.created_at = source["created_at"];
	    }
	}

}

