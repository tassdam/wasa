import { createRouter, createWebHashHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import HomeView from "../views/HomeView.vue";
import SearchPeopleView from "../views/SearchPeopleView.vue";
import ChatView from "../views/ChatView.vue";

const routes = [
  { path: "/", component: LoginView },
  { path: "/home", component: HomeView },
  { path: "/search", component: SearchPeopleView },
  { path: "/conversations/:uuid", name: "ChatView", component: ChatView, props: true },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
