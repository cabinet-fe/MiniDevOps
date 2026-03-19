import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { REPO_AUTH_TYPES } from "@/lib/constants";

interface ProjectPayload {
  name: string;
  description: string;
  repo_url: string;
  repo_auth_type: string;
  repo_username: string;
  repo_password: string;
  max_artifacts: number;
}

interface ProjectDetail extends Omit<ProjectPayload, "repo_password"> {
  id: number;
}

const DEFAULT_FORM: ProjectPayload = {
  name: "",
  description: "",
  repo_url: "",
  repo_auth_type: "none",
  repo_username: "",
  repo_password: "",
  max_artifacts: 5,
};

export function ProjectFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = !!id;

  const [loading, setLoading] = useState(isEdit);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [form, setForm] = useState<ProjectPayload>(DEFAULT_FORM);

  useEffect(() => {
    if (!isEdit || !id) {
      return;
    }

    const fetchProject = async () => {
      setLoading(true);
      try {
        const res = await api.get<ProjectDetail>(`/projects/${id}`);
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || "加载项目失败");
        }
        setForm({
          name: res.data.name || "",
          description: res.data.description || "",
          repo_url: res.data.repo_url || "",
          repo_auth_type: res.data.repo_auth_type || "none",
          repo_username: res.data.repo_username || "",
          repo_password: "",
          max_artifacts: res.data.max_artifacts || 5,
        });
      } catch (err) {
        const message = err instanceof Error ? err.message : "加载项目失败";
        setError(message);
        toast.error(message);
      } finally {
        setLoading(false);
      }
    };

    fetchProject();
  }, [id, isEdit]);

  const setField = <K extends keyof ProjectPayload>(key: K, value: ProjectPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const validate = () => {
    if (!form.name.trim()) return "请输入项目名称";
    if (!form.repo_url.trim()) return "请输入仓库地址";
    if (form.max_artifacts < 1) return "构建产物保留数量必须大于 0";
    if (!isEdit && form.repo_auth_type !== "none" && !form.repo_password.trim()) {
      return "请填写仓库认证信息";
    }
    return "";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    setError("");
    setSubmitting(true);

    try {
      const payload: ProjectPayload = {
        name: form.name.trim(),
        description: form.description.trim(),
        repo_url: form.repo_url.trim(),
        repo_auth_type: form.repo_auth_type,
        repo_username: form.repo_auth_type === "none" ? "" : form.repo_username.trim(),
        repo_password: form.repo_auth_type === "none" ? "" : form.repo_password,
        max_artifacts: form.max_artifacts,
      };

      if (isEdit && id) {
        const res = await api.put<ProjectDetail>(`/projects/${id}`, payload);
        if (res.code !== 0) {
          throw new Error(res.message || "更新项目失败");
        }
        toast.success("项目已更新");
        navigate(`/projects/${id}`, { replace: true });
        return;
      }

      const res = await api.post<ProjectDetail>("/projects", payload);
      if (res.code !== 0 || !res.data) {
        throw new Error(res.message || "创建项目失败");
      }
      toast.success("项目创建成功");
      navigate(`/projects/${res.data.id}`, { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : "提交失败";
      setError(message);
      toast.error(message);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{isEdit ? "编辑项目" : "新建项目"}</h1>
          <p className="mt-1 text-sm text-zinc-500">
            {isEdit ? "更新项目仓库与构建配置" : "创建新的构建项目并配置仓库信息"}
          </p>
        </div>
        <Link to={isEdit && id ? `/projects/${id}` : "/projects"}>
          <Button variant="outline">返回</Button>
        </Link>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>项目配置</CardTitle>
          <CardDescription>带 * 的字段为必填项</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                {error}
              </div>
            )}

            <div className="grid gap-5 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="name">项目名称 *</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => setField("name", e.target.value)}
                  placeholder="例如：buildflow-web"
                  maxLength={100}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="max-artifacts">构建产物保留数量 *</Label>
                <Input
                  id="max-artifacts"
                  type="number"
                  min={1}
                  value={form.max_artifacts}
                  onChange={(e) =>
                    setField("max_artifacts", Math.max(1, Number(e.target.value) || 1))
                  }
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="repo-url">仓库地址 *</Label>
              <Input
                id="repo-url"
                value={form.repo_url}
                onChange={(e) => setField("repo_url", e.target.value)}
                placeholder="https://github.com/org/repo.git"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">描述</Label>
              <Textarea
                id="description"
                value={form.description}
                onChange={(e) => setField("description", e.target.value)}
                placeholder="简要描述该项目用途"
                rows={3}
                maxLength={500}
              />
            </div>

            <div className="grid gap-5 md:grid-cols-2">
              <div className="space-y-2">
                <Label>仓库认证方式</Label>
                <Select
                  value={form.repo_auth_type}
                  onValueChange={(value) => {
                    if (value === "none") {
                      setForm((prev) => ({
                        ...prev,
                        repo_auth_type: value,
                        repo_username: "",
                        repo_password: "",
                      }));
                      return;
                    }
                    setField("repo_auth_type", value);
                  }}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {REPO_AUTH_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {form.repo_auth_type !== "none" && (
                <div className="space-y-2">
                  <Label htmlFor="repo-username">仓库用户名</Label>
                  <Input
                    id="repo-username"
                    value={form.repo_username}
                    onChange={(e) => setField("repo_username", e.target.value)}
                    placeholder="可选，部分仓库类型需要"
                  />
                </div>
              )}
            </div>

            {form.repo_auth_type !== "none" && (
              <div className="space-y-2">
                <Label htmlFor="repo-password">
                  {isEdit ? "仓库密码 / Token（留空表示不变）" : "仓库密码 / Token *"}
                </Label>
                <Input
                  id="repo-password"
                  type="password"
                  value={form.repo_password}
                  onChange={(e) => setField("repo_password", e.target.value)}
                  placeholder={isEdit ? "如不修改可留空" : "请输入仓库密码或访问 Token"}
                />
              </div>
            )}

            <div className="flex justify-end gap-2 pt-2">
              <Link to={isEdit && id ? `/projects/${id}` : "/projects"}>
                <Button type="button" variant="outline">
                  取消
                </Button>
              </Link>
              <Button type="submit" disabled={submitting}>
                {submitting ? (isEdit ? "保存中..." : "创建中...") : isEdit ? "保存" : "创建项目"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
