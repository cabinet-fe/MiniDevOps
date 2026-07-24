import { createApp } from "vue";
import { createPinia } from "pinia";
import { vLoading } from "@veltra/desktop";
import { loadTheme, setTheme } from "@veltra/styles/theme";
import "@veltra/styles/normalize";
import "@veltra/styles/transitions";
import "@veltra/desktop/components/message/style.js";
import "@veltra/desktop/components/loading/style.js";

import App from "./App.vue";
import router from "./router";
import { setOnAuthExpired } from "./api/http";
import { useAuthStore } from "./stores/auth";
import { bedrockTheme } from "./theme/bedrock";

// setTheme("light") sets html[data-theme=light] so Veltra injects light component CSS vars.
// loadTheme alone does not flip the attribute.
setTheme("light");
loadTheme(bedrockTheme);

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

// Resolver covers U* components only; directives must be registered manually.
app.directive("loading", vLoading);

setOnAuthExpired(() => {
  const auth = useAuthStore();
  auth.clearSession();
  if (router.currentRoute.value.name !== "login") {
    void router.replace({ name: "login" });
  }
});

app.mount("#app");
