import { apiData, http } from "./http";
import type { Dictionary, OperationLog, PageResult, RbacResource, Role, User } from "./types";

export type ListQuery = Record<string, string | number | boolean | undefined | null>;

function toQuery(params?: ListQuery): Record<string, string | number | boolean> {
  const out: Record<string, string | number | boolean> = {};
  if (!params) return out;
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue;
    out[k] = v;
  }
  return out;
}

export async function listUsers(params?: ListQuery): Promise<PageResult<User>> {
  return apiData(http.get("/users", { query: toQuery(params) }));
}

export async function createUser(body: Record<string, unknown>): Promise<User> {
  return apiData(http.post("/users", body));
}

export async function updateUser(id: number, body: Record<string, unknown>): Promise<User> {
  return apiData(http.put(`/users/${id}`, body));
}

export async function deleteUser(id: number): Promise<void> {
  await apiData(http.delete(`/users/${id}`));
}

export async function listRoles(params?: ListQuery): Promise<PageResult<Role>> {
  return apiData(http.get("/roles", { query: toQuery(params) }));
}

export async function getRole(id: number): Promise<Role> {
  return apiData(http.get(`/roles/${id}`));
}

export async function createRole(body: Record<string, unknown>): Promise<Role> {
  return apiData(http.post("/roles", body));
}

export async function updateRole(id: number, body: Record<string, unknown>): Promise<Role> {
  return apiData(http.put(`/roles/${id}`, body));
}

export async function setRolePermissions(id: number, permissions: string[]): Promise<Role> {
  return apiData(http.put(`/roles/${id}/permissions`, { permissions }));
}

export async function deleteRole(id: number): Promise<void> {
  await apiData(http.delete(`/roles/${id}`));
}

export async function listResources(): Promise<{ items: RbacResource[] }> {
  return apiData(http.get("/rbac/resources"));
}

export async function createResource(body: Record<string, unknown>): Promise<RbacResource> {
  return apiData(http.post("/rbac/resources", body));
}

export async function updateResource(
  id: number,
  body: Record<string, unknown>,
): Promise<RbacResource> {
  return apiData(http.put(`/rbac/resources/${id}`, body));
}

export async function deleteResource(id: number): Promise<void> {
  await apiData(http.delete(`/rbac/resources/${id}`));
}

export async function listMenus(): Promise<{ items: RbacResource[] }> {
  return apiData(http.get("/menus"));
}

export async function updateMenu(id: number, body: Record<string, unknown>): Promise<RbacResource> {
  return apiData(http.put(`/menus/${id}`, body));
}

export async function updateMenuIcon(
  id: number,
  iconBase64: string,
  iconMime?: string,
): Promise<RbacResource> {
  return apiData(http.put(`/menus/${id}/icon`, { icon_base64: iconBase64, icon_mime: iconMime }));
}

export async function listDictionaries(params?: ListQuery): Promise<PageResult<Dictionary>> {
  return apiData(http.get("/dictionaries", { query: toQuery(params) }));
}

export async function getDictionary(id: number): Promise<Dictionary> {
  return apiData(http.get(`/dictionaries/${id}`));
}

export async function createDictionary(body: Record<string, unknown>): Promise<Dictionary> {
  return apiData(http.post("/dictionaries", body));
}

export async function updateDictionary(
  id: number,
  body: Record<string, unknown>,
): Promise<Dictionary> {
  return apiData(http.put(`/dictionaries/${id}`, body));
}

export async function deleteDictionary(id: number): Promise<void> {
  await apiData(http.delete(`/dictionaries/${id}`));
}

export async function listOperationLogs(params?: ListQuery): Promise<PageResult<OperationLog>> {
  return apiData(http.get("/operation-logs", { query: toQuery(params) }));
}
