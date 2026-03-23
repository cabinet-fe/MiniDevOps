const API_BASE = "/api/v1";

interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

let getToken: () => string | null = () => localStorage.getItem("access_token");

export function setTokenGetter(fn: () => string | null) {
  getToken = fn;
}

function buildHeaders(isFormData = false): Record<string, string> {
  const headers: Record<string, string> = {};
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;
  if (!isFormData) headers["Content-Type"] = "application/json";
  return headers;
}

async function authorizedFetch(
  method: string,
  path: string,
  body?: unknown,
  isFormData = false,
): Promise<Response> {
  const headers = buildHeaders(isFormData);

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    credentials: "omit",
    body: isFormData ? (body as FormData) : body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    const shouldTryRefresh = path !== "/auth/login" && path !== "/auth/refresh";
    if (shouldTryRefresh) {
      const refreshToken = localStorage.getItem("refresh_token");
      if (refreshToken) {
        const refreshRes = await fetch(`${API_BASE}/auth/refresh`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (refreshRes.ok) {
          const data = await refreshRes.json();
          if (data.data?.access_token) {
            localStorage.setItem("access_token", data.data.access_token);
            if (data.data.refresh_token) {
              localStorage.setItem("refresh_token", data.data.refresh_token);
            }
            headers["Authorization"] = `Bearer ${data.data.access_token}`;
            const retryRes = await fetch(`${API_BASE}${path}`, {
              method,
              headers,
              body: isFormData ? (body as FormData) : body ? JSON.stringify(body) : undefined,
            });
            return retryRes;
          }
        }
      }
    }

    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    if (window.location.pathname !== "/login") {
      window.location.assign("/login");
    }
    throw new Error("Unauthorized");
  }

  return res;
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  isFormData = false,
): Promise<ApiResponse<T>> {
  const res = await authorizedFetch(method, path, body, isFormData);
  return res.json();
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  put: <T>(path: string, body?: unknown) => request<T>("PUT", path, body),
  delete: <T>(path: string) => request<T>("DELETE", path),
  upload: <T>(path: string, formData: FormData) => request<T>("POST", path, formData, true),
  getText: async (path: string) => {
    const res = await authorizedFetch("GET", path);
    return res.text();
  },
  download: async (path: string) => {
    const res = await authorizedFetch("GET", path);
    return res.blob();
  },
};

export type { ApiResponse, PaginatedData };
