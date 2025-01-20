import { createRouter, createWebHashHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import HomeView from "../views/HomeView.vue";
import SearchPeopleView from "../views/SearchPeopleView.vue";
import ChatView from "../views/ChatView.vue";
import ProfileView from "../views/ProfileView.vue";
import GroupsView from "../views/GroupsView.vue"
import GroupCreateView from "../views/GroupCreateView.vue"
import GroupEditView from "../views/GroupEditView.vue"

const routes = [
  { path: "/", component: LoginView },
  { path: "/home", component: HomeView },
  { path: "/search", component: SearchPeopleView },
  { path: "/conversations/:uuid", name: "ChatView", component: ChatView, props: true },
  { path: "/me", component: ProfileView},
  { path: "/groups", component: GroupsView},
  { path: "/new-group", component: GroupCreateView},
  { path: "/groups/:uuid", name: "GroupEditView", component: GroupEditView, props: true}
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
