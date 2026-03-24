import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useAuthStore } from "@/stores/auth-store";
import { cn } from "@/lib/utils";
import { loadSavedCredentials, persistCredentials } from "@/lib/login-persist";
import { useTheme } from "next-themes";
import { Sun, Moon } from "lucide-react";
import { Button } from "@/components/ui/button";

export function LoginPage() {
  const saved = loadSavedCredentials();
  const navigate = useNavigate();
  const { login, isAuthenticated, fetchMe, token } = useAuthStore();
  const [username, setUsername] = useState(saved.username);
  const [password, setPassword] = useState(saved.password);
  const [remember, setRemember] = useState(saved.remember);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { theme, setTheme } = useTheme();

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/", { replace: true });
      return;
    }

    if (!token) {
      return;
    }

    let active = true;
    fetchMe().then(() => {
      if (active && useAuthStore.getState().isAuthenticated) {
        navigate("/", { replace: true });
      }
    });
    return () => {
      active = false;
    };
  }, [isAuthenticated, navigate, fetchMe, token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password) return;
    setError("");
    setLoading(true);
    try {
      await login(username, password);
      persistCredentials(remember, username, password);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "错误：登录失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden bg-slate-50 dark:bg-[#0a0a0a] text-slate-900 dark:text-[#33ff00] font-mono selection:bg-slate-900 selection:text-slate-50 dark:selection:bg-[#33ff00] dark:selection:text-[#0a0a0a]">
      {/* Theme Toggle */}
      <div className="absolute top-4 right-4 z-[60]">
        <Button
          variant="ghost"
          size="icon"
          className="size-9 text-slate-500 dark:text-[#1f521f] hover:bg-slate-200 dark:hover:bg-[#1f521f]/20 hover:text-slate-900 dark:hover:text-[#33ff00] border border-transparent hover:border-slate-300 dark:hover:border-[#33ff00]/30 transition-all"
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          title="切换主题"
        >
          <Sun className="size-5 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute size-5 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
          <span className="sr-only">切换主题</span>
        </Button>
      </div>

      {/* CRT Scanline Overlay */}
      <div 
        className="pointer-events-none fixed inset-0 z-50 opacity-[0.03] dark:opacity-[0.12]"
        style={{
          background: "linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.25) 50%), linear-gradient(90deg, rgba(255, 0, 0, 0.06), rgba(0, 255, 0, 0.02), rgba(0, 0, 255, 0.06))",
          backgroundSize: "100% 2px, 3px 100%"
        }}
        aria-hidden
      />

      <style>{`
        .terminal-glow, .terminal-glow-amber, .terminal-glow-red {
          text-shadow: none;
        }
        .dark .terminal-glow {
          text-shadow: 0 0 5px rgba(51, 255, 0, 0.5);
        }
        .dark .terminal-glow-amber {
          text-shadow: 0 0 5px rgba(255, 176, 0, 0.5);
        }
        .dark .terminal-glow-red {
          text-shadow: 0 0 5px rgba(255, 51, 51, 0.5);
        }
        /* Light mode autofill */
        input:-webkit-autofill,
        input:-webkit-autofill:hover, 
        input:-webkit-autofill:focus, 
        input:-webkit-autofill:active{
            -webkit-box-shadow: 0 0 0 30px #f8fafc inset !important;
            -webkit-text-fill-color: #0f172a !important;
            caret-color: #0f172a !important;
        }
        /* Dark mode autofill */
        .dark input:-webkit-autofill,
        .dark input:-webkit-autofill:hover, 
        .dark input:-webkit-autofill:focus, 
        .dark input:-webkit-autofill:active{
            -webkit-box-shadow: 0 0 0 30px #0a0a0a inset !important;
            -webkit-text-fill-color: #33ff00 !important;
            caret-color: #33ff00 !important;
        }
      `}</style>

      <div className="z-10 w-full max-w-2xl p-4 sm:p-6 md:p-8 terminal-glow">
        <div className="mb-6 hidden sm:block whitespace-pre text-[10px] sm:text-xs md:text-sm leading-tight text-slate-900 dark:text-[#33ff00]">
{`   _         _ _    _  __ _                
  | |__ _  _(_) |__| |/ _| |_____ __ __    
  | '_ \\ || | | / _\` |  _| / _ \\ V  V /    
  |_.__/\\_,_|_|_\\__,_|_| |_\\___/\\_/\\_/     

> 系统平台：BuildFlow CI/CD 持续发布与部署系统
> 核心引擎：调度中心 (Scheduler) 已在线
> 当前状态：等待认证登入以接管构建管道
> 核心版本：${import.meta.env.VITE_APP_VERSION || "v1.0.0-dev"}
================================================`}
        </div>

        {/* Mobile Header */}
        <div className="mb-6 block sm:hidden whitespace-pre text-xs leading-tight text-slate-900 dark:text-[#33ff00]">
{`> 系统平台：BuildFlow CI/CD 部署系统
> 核心版本：${import.meta.env.VITE_APP_VERSION || "v1.0.0-dev"}
> 等待认证登入以接管构建管道
===========================`}
        </div>

        <div className="border border-slate-300 dark:border-[#1f521f] bg-slate-50 dark:bg-[#0a0a0a] p-1 shadow-sm dark:shadow-none">
          <div className="border border-dashed border-slate-400 dark:border-[#1f521f] p-4 sm:p-6 md:p-8 bg-white dark:bg-transparent">
            <h1 className="mb-8 font-bold uppercase tracking-widest text-amber-700 dark:text-[#ffb000] terminal-glow-amber">
              +--- 登入安全构建调度环境 ---+
            </h1>

            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <div className="mb-4 border border-red-500 dark:border-[#ff3333] bg-red-50 dark:bg-transparent p-3 text-red-600 dark:text-[#ff3333] terminal-glow-red uppercase text-sm" role="alert">
                  [{error}]
                </div>
              )}

              <div className="grid grid-cols-1 sm:grid-cols-[max-content_1fr] gap-x-4 gap-y-4 sm:gap-y-6 sm:items-center">
                <label htmlFor="username" className="text-sm md:text-base uppercase flex items-center shrink-0">
                  认证用户 @buildflow:~$ 
                </label>
                <div className="flex-1">
                  <input
                    id="username"
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    autoComplete="username"
                    autoFocus
                    spellCheck={false}
                    className="w-full bg-transparent border-none outline-none text-slate-900 dark:text-[#33ff00] focus:ring-0 p-0 text-sm md:text-base selection:bg-slate-900 selection:text-slate-50 dark:selection:bg-[#33ff00] dark:selection:text-[#0a0a0a] caret-slate-900 dark:caret-[#33ff00]"
                  />
                </div>

                <label htmlFor="password" className="text-sm md:text-base uppercase flex items-center shrink-0">
                  访问凭据:
                </label>
                <div className="flex-1">
                  <input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    autoComplete="current-password"
                    className="w-full bg-transparent border-none outline-none text-slate-900 dark:text-[#33ff00] focus:ring-0 p-0 text-sm md:text-base tracking-[0.2em] sm:tracking-[0.3em] selection:bg-slate-900 selection:text-slate-50 dark:selection:bg-[#33ff00] dark:selection:text-[#0a0a0a] caret-slate-900 dark:caret-[#33ff00]"
                  />
                </div>
              </div>

              <div className="mt-6 flex items-center gap-3 text-sm text-slate-500 dark:text-[#1f521f] transition-colors hover:text-slate-900 dark:hover:text-[#33ff00]">
                <button
                  type="button"
                  onClick={() => setRemember(!remember)}
                  className="flex items-center gap-2 outline-none focus:text-amber-700 dark:focus:text-[#ffb000]"
                >
                  <span className="font-bold">[{remember ? "X" : " "}]</span>
                  <span className="uppercase tracking-wider">保持登录状态 (7天)_</span>
                </button>
              </div>

              <div className="pt-6 border-t border-dashed border-slate-300 dark:border-[#1f521f] flex flex-col sm:flex-row items-start sm:items-center gap-4">
                <button
                  type="submit"
                  disabled={loading}
                  className="group relative inline-flex items-center outline-none"
                >
                  <span className="mr-3 text-amber-700 dark:text-[#ffb000]">&gt;</span>
                  <span className={cn(
                    "px-3 py-1.5 uppercase font-bold tracking-widest transition-all",
                    loading 
                      ? "text-slate-400 dark:text-[#1f521f]" 
                      : "bg-slate-100 dark:bg-[#0a0a0a] text-slate-900 dark:text-[#33ff00] border border-slate-900 dark:border-[#33ff00] group-hover:bg-slate-900 group-hover:text-white dark:group-hover:bg-[#33ff00] dark:group-hover:text-[#0a0a0a] group-focus:bg-slate-900 group-focus:text-white dark:group-focus:bg-[#33ff00] dark:group-focus:text-[#0a0a0a]"
                  )}>
                    {loading ? "[ 初始化通讯与密钥分发中... ]" : "[ 登入 BuildFlow 流水线 ]"}
                  </span>
                </button>
              </div>
            </form>
          </div>
        </div>

        <div className="mt-8 text-xs text-slate-500 dark:text-[#1f521f] text-center space-y-1">
          <p>BUILDFLOW 集中式构建引擎与工程自动化平台。</p>
          <p>当前终端的资源配置、构建事件与持续交付行为均受 RBAC 审计追踪。</p>
        </div>
      </div>
    </div>
  );
}
