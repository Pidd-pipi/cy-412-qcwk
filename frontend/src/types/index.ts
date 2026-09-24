export type UserRole='resident'|'staff'|'admin'; export type RepairStatus='pending'|'assigned'|'processing'|'done'|'closed';
/** 公告发布范围：全小区 / 指定楼栋 / 指定单元 */
export type AnnouncementScope='all'|'building'|'unit';
export interface User{id:number;phone:string;nickname:string;avatar:string;role:UserRole;building:string;unit:string;room:string;created_at:string}; export interface Repair{id:number;user_id:number;title:string;description:string;type:string;images:string;status:RepairStatus;handler_id?:number;user?:User;handler?:User;rating:number;created_at:string;updated_at:string}; export interface Payment{id:number;user_id:number;fee_type:string;amount:number;month:string;status:string;paid_at?:string;created_at:string}; export interface Announcement{id:number;title:string;content:string;category:string;publisher_id:number;publisher?:User;publish_at:string;top:boolean;read_count:number;scope:AnnouncementScope;building:string;unit:string;
  /** 紧急公告：当前住户是否已确认 */
  confirmed?:boolean;
  /** 紧急公告：物业视角的确认人数 */
  confirmed_count?:number;
  /** 紧急公告：物业视角的未确认住户名单 */
  unconfirmed_users?:User[]};
export interface ScopeOptions{buildings:string[];units:string[]};
export interface ApiResponse<T>{code:number;message:string;data:T}
