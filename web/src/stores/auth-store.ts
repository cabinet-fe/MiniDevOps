import { create } from "zustand";
import { api } from "@/lib/api";
import { encryptLoginPassword } from "@/lib/login-crypto";

interface User {
  id: number;
  username: string;
  display_name: string;
  role: string;
  email: string;
  avatar: string;
  is_active: boolean;
}

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchMe: () => Promise<void>;
  setToken: (token: string) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem("access_token"),
  user: null,
  isAuthenticated: !!localStorage.getItem("access_token"),

  login: async (username, password) => {
    const password_cipher = await encryptLoginPassword(password);
    const res = await api.post<{ access_token: string; refresh_token: string; user: User }>(
      "/auth/login",
      {
        username,
        password_cipher,
      },
    );
    if (res.code !== 0) throw new Error(res.message);
    localStorage.setItem("access_token", res.data!.access_token);
    localStorage.setItem("refresh_token", res.data!.refresh_token);
    set({ token: res.data!.access_token, user: res.data!.user, isAuthenticated: true });
  },

  logout: async () => {
    try {
      await api.post("/auth/logout");
    } catch {}
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    set({ token: null, user: null, isAuthenticated: false });
  },

  fetchMe: async () => {
    const token = get().token ?? localStorage.getItem("access_token");
    if (!token) {
      set({ token: null, user: null, isAuthenticated: false });
      return;
    }

    try {
      const res = await api.get<User>("/auth/me");
      if (res.code === 0) set({ user: res.data!, isAuthenticated: true });
    } catch {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      set({ token: null, user: null, isAuthenticated: false });
    }
  },

  setToken: (token) => {
    localStorage.setItem("access_token", token);
    set({ token, isAuthenticated: true });
  },
}));

export type { User };
