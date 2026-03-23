import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { Factory, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";
import { cn } from "@/lib/utils";
import { loadSavedCredentials, persistCredentials } from "@/lib/login-persist";

export function LoginPage() {
  const saved = loadSavedCredentials();
  const navigate = useNavigate();
  const { login, isAuthenticated, fetchMe, token } = useAuthStore();
  const [username, setUsername] = useState(saved.username);
  const [password, setPassword] = useState(saved.password);
  const [remember, setRemember] = useState(saved.remember);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

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
    setError("");
    setLoading(true);
    try {
      await login(username, password);
      persistCredentials(remember, username, password);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen flex-col overflow-hidden bg-[#0c0c0e] text-zinc-100 md:flex-row">
      {/* 顶栏：细警示条纹带 */}
      <div
        className="pointer-events-none absolute left-0 right-0 top-0 z-[2] h-1.5"
        style={{
          backgroundImage: `repeating-linear-gradient(
            -45deg,
            #0a0a0a,
            #0a0a0a 5px,
            rgba(234, 179, 8, 0.55) 5px,
            rgba(234, 179, 8, 0.55) 10px
          )`,
        }}
        aria-hidden
      />

      {/* 背景：暖锈底 + 钢质感网格 + 底影 */}
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          background: `
            radial-gradient(ellipse 85% 55% at 50% -8%, rgba(194, 65, 12, 0.14), transparent 52%),
            radial-gradient(ellipse 70% 45% at 100% 50%, rgba(63, 63, 70, 0.35), transparent 55%),
            linear-gradient(168deg, #161311 0%, #0a0a0c 42%, #0e0d0c 100%)
          `,
        }}
      />
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.42]"
        style={{
          backgroundImage: `
            linear-gradient(95deg, transparent 0%, rgba(161, 161, 170, 0.06) 38%, transparent 72%),
            repeating-linear-gradient(
              0deg,
              transparent,
              transparent 56px,
              rgba(82, 82, 91, 0.11) 56px,
              rgba(82, 82, 91, 0.11) 57px
            ),
            repeating-linear-gradient(
              90deg,
              transparent,
              transparent 56px,
              rgba(63, 63, 70, 0.08) 56px,
              rgba(63, 63, 70, 0.08) 57px
            )
          `,
        }}
      />
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_bottom,rgba(0,0,0,0.45),transparent_60%)]" />

      {/* 左侧：车间品牌区 + 侧边警示边条 */}
      <aside className="relative z-[1] flex flex-none flex-col justify-between border-b border-zinc-700/50 px-8 py-10 md:w-[42%] md:border-b-0 md:border-r md:py-14 lg:px-14">
        <div
          className="pointer-events-none absolute bottom-0 left-0 top-0 w-1 md:w-1.5"
          style={{
            backgroundImage: `repeating-linear-gradient(
              45deg,
              #0a0a0a,
              #0a0a0a 4px,
              rgba(234, 179, 8, 0.28) 4px,
              rgba(234, 179, 8, 0.28) 8px
            )`,
          }}
          aria-hidden
        />
        <div>
          <div className="flex items-center gap-3 font-mono text-sm tracking-[0.2em] text-orange-500/95">
            <Factory className="size-5 shrink-0 text-orange-400" aria-hidden />
            <span>BUILDFLOW</span>
          </div>
          <p className="mt-3 font-mono text-[10px] uppercase tracking-[0.35em] text-zinc-500">
            sector · access
          </p>
          <h1 className="mt-8 max-w-md font-mono text-3xl font-medium leading-tight tracking-tight text-zinc-50 md:text-4xl">
            流水线
            <span className="text-orange-600">.</span>
            <br />
            构建与部署
          </h1>
          <p className="mt-4 max-w-sm text-sm leading-relaxed text-zinc-400">
            受控发布、可追溯构建日志与多环境编排。登录以进入控制台。
          </p>
        </div>
        <div className="mt-12 hidden font-mono text-[10px] leading-relaxed text-zinc-500 md:block">
          <div className="border-l-2 border-orange-700/55 pl-3">
            <p>PLANT: ops-ready</p>
            <p className="mt-1 text-zinc-400">AUTH: JWT · WS: channels</p>
          </div>
        </div>
      </aside>

      {/* 右侧：表单 — 钢板面板 */}
      <main className="relative z-[1] flex flex-1 items-center justify-center px-6 py-12 md:px-12">
        <div className="w-full max-w-[400px]">
          <div className="mb-8 md:hidden">
            <p className="font-mono text-xs tracking-widest text-orange-500/85">BUILDFLOW</p>
            <p className="mt-1 text-lg font-medium text-zinc-100">登录</p>
          </div>

          <div
            className={cn(
              "relative rounded-lg border-2 border-zinc-600/70 bg-[#141416]/95 p-8",
              "shadow-[inset_0_1px_0_rgba(255,255,255,0.06),0_20px_50px_rgba(0,0,0,0.55)]",
              "backdrop-blur-sm",
            )}
          >
            {/* 角码装饰 */}
            <span
              className="pointer-events-none absolute left-2 top-2 size-3 border-l-2 border-t-2 border-zinc-500/80"
              aria-hidden
            />
            <span
              className="pointer-events-none absolute right-2 top-2 size-3 border-r-2 border-t-2 border-zinc-500/80"
              aria-hidden
            />
            <span
              className="pointer-events-none absolute bottom-2 left-2 size-3 border-b-2 border-l-2 border-zinc-500/80"
              aria-hidden
            />
            <span
              className="pointer-events-none absolute bottom-2 right-2 size-3 border-b-2 border-r-2 border-zinc-500/80"
              aria-hidden
            />
            <form onSubmit={handleSubmit} className="space-y-5">
              {error && (
                <div
                  className="rounded border border-red-600/40 bg-red-950/50 px-3 py-2.5 font-mono text-xs text-red-300 shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]"
                  role="alert"
                >
                  {error}
                </div>
              )}
              <div className="space-y-2">
                <Label
                  htmlFor="username"
                  className="font-mono text-xs uppercase tracking-wider text-zinc-400"
                >
                  用户名
                </Label>
                <Input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="operator"
                  required
                  autoComplete="username"
                  className="h-11 border-zinc-600/90 bg-zinc-900/90 font-mono text-sm text-zinc-50 shadow-[inset_0_2px_4px_rgba(0,0,0,0.35)] placeholder:text-zinc-500 focus-visible:border-orange-600/60 focus-visible:ring-orange-600/25"
                />
              </div>
              <div className="space-y-2">
                <Label
                  htmlFor="password"
                  className="font-mono text-xs uppercase tracking-wider text-zinc-400"
                >
                  密码
                </Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  autoComplete="current-password"
                  className="h-11 border-zinc-600/90 bg-zinc-900/90 font-mono text-sm text-zinc-50 shadow-[inset_0_2px_4px_rgba(0,0,0,0.35)] placeholder:text-zinc-500 focus-visible:border-orange-600/60 focus-visible:ring-orange-600/25"
                />
              </div>

              <label className="flex cursor-pointer items-start gap-3 rounded border border-transparent px-0 py-1 hover:border-zinc-600/50">
                <input
                  type="checkbox"
                  checked={remember}
                  onChange={(e) => setRemember(e.target.checked)}
                  className="mt-0.5 size-4 shrink-0 rounded border-zinc-500 bg-zinc-900 accent-orange-600"
                />
                <span className="text-sm leading-snug text-zinc-400">
                  记住账号密码
                  <span className="mt-0.5 block font-mono text-[11px] text-zinc-500">
                    凭据保存在本机浏览器，请勿在公共设备勾选
                  </span>
                </span>
              </label>

              <Button
                type="submit"
                disabled={loading}
                className={cn(
                  "h-11 w-full font-mono text-sm tracking-wide",
                  "border border-orange-900/60 bg-orange-700 text-zinc-50 shadow-[0_4px_0_rgba(124,45,18,0.85)] hover:bg-orange-600 active:translate-y-px active:shadow-[0_2px_0_rgba(124,45,18,0.85)]",
                  loading && "opacity-80",
                )}
              >
                {loading ? (
                  <>
                    <Loader2 className="size-4 animate-spin" />
                    校验中…
                  </>
                ) : (
                  "进入控制台"
                )}
              </Button>
            </form>
          </div>
        </div>
      </main>
    </div>
  );
}
