import { createApp } from "vue";
import { createPinia } from "pinia";
import { loadTheme, heroLightTheme } from "@veltra/styles/theme";
import "@veltra/styles/normalize";
import "@veltra/styles/transitions";
import "@veltra/desktop/components/message/style.js";

import App from "./App.vue";
import router from "./router";
import { setOnAuthExpired } from "./api/http";
import { useAuthStore } from "./stores/auth";

// Inject --u-* design tokens before mount (required for Veltra component look).
loadTheme(heroLightTheme);

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

setOnAuthExpired(() => {
  const auth = useAuthStore();
  auth.clearSession();
  if (router.currentRoute.value.name !== "login") {
    void router.replace({ name: "login" });
  }
});

app.mount("#app");
