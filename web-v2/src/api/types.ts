export interface ApiEnvelope<T = unknown> {
  code: number;
  message: string;
  data?: T;
  request_id?: string;
}

export interface User {
  id: number;
  username: string;
  display_name: string;
  email: string;
  avatar: string;
  is_active: boolean;
  is_super_admin: boolean;
  role_ids?: number[];
}

export interface MenuNode {
  path: string;
  title: string;
  route?: string;
  icon?: string;
  sort?: number;
  children?: MenuNode[];
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  user: User;
  permissions?: string[];
  menus?: MenuNode[];
}

export interface MeResponse {
  user: User;
  permissions: string[];
  menus: MenuNode[];
}

export interface PageResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface RolePermission {
  id: number;
  role_id: number;
  permission: string;
}

export interface Role {
  id: number;
  name: string;
  code: string;
  description: string;
  permissions?: RolePermission[];
}

export interface MenuMetadata {
  id: number;
  resource_id: number;
  title: string;
  route: string;
  icon_base64?: string;
  icon_mime?: string;
}

export interface RbacResource {
  id: number;
  path: string;
  type: "menu" | "page" | "action" | "card";
  parent_id?: number | null;
  enabled: boolean;
  sort_key: number;
  menu_metadata?: MenuMetadata;
  children?: RbacResource[];
}

export interface Dictionary {
  id: number;
  name: string;
  code: string;
  description: string;
  items?: DictItem[];
}

export interface DictItem {
  id?: number;
  dictionary_id?: number;
  label: string;
  value: string;
  sort_order?: number;
  enabled?: boolean;
}

export interface OperationLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details: string;
  ip_address: string;
  created_at: string;
}

export interface Credential {
  id: number;
  name: string;
  type: "password" | "token" | "ssh_key" | "api_key" | string;
  username: string;
  description: string;
  has_secret: boolean;
  has_passphrase?: boolean;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface Repository {
  id: number;
  name: string;
  description: string;
  tags: string;
  repo_url: string;
  default_branch: string;
  auth_type: string;
  credential_id?: number | null;
  webhook_type?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface Server {
  id: number;
  name: string;
  host: string;
  port: number;
  os_type: string;
  username: string;
  auth_type: string;
  credential_id?: number | null;
  agent_url?: string;
  agent_credential_id?: number | null;
  description: string;
  tags: string;
  status: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface DeployTarget {
  id?: number;
  build_job_id?: number;
  server_id?: number | null;
  remote_path: string;
  method: string;
  post_deploy_script?: string;
  sort_order: number;
}

export interface BuildJob {
  id: number;
  repository_id: number;
  name: string;
  description: string;
  enabled: boolean;
  branch_policy: string;
  branch: string;
  shallow_clone: boolean;
  build_script_type: string;
  build_script: string;
  work_dir: string;
  output_dir: string;
  cache_paths: string;
  env_var_names?: string[];
  trigger_manual: boolean;
  trigger_webhook: boolean;
  trigger_cron: boolean;
  cron_expression: string;
  cron_timezone: string;
  max_artifacts: number;
  artifact_format: string;
  agent_trigger_event: string;
  deploy_targets?: DeployTarget[];
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface BuildDeployAttempt {
  id: number;
  build_run_id: number;
  batch_no: number;
  deploy_target_id?: number | null;
  status: string;
  error_message?: string;
  created_at: string;
}

export interface BuildRun {
  id: number;
  build_job_id: number;
  build_number: number;
  status: string;
  stage: string;
  trigger_type: string;
  triggered_by: number;
  branch: string;
  commit_hash: string;
  commit_message: string;
  log_path?: string;
  artifact_path?: string;
  distribution_summary: string;
  snapshot_json?: string;
  error_message?: string;
  created_at: string;
  deploy_attempts?: BuildDeployAttempt[];
}
