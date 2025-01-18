<script>
export default {
  data() {
    return {
      username: "", 
      errormsg: null,
      loading: false,
      conversations: [] 
    };
  },
  methods: {
    async loadConversations() {
      this.errormsg = null;
      this.loading = true;

      try {
        // Get the token from localStorage
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        const response = await this.$axios.get("/users/me/conversations", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
        this.conversations = response.data || [];
      } catch (error) {
        console.error("Error loading conversations:", error);
        this.errormsg = "Failed to load conversations. Please try again.";
      } finally {
        this.loading = false;
      }
    },
    viewConversation(conversationId) {   
      this.$router.push({
        path: `/conversations/${conversationId}`
      });
    },
    refresh() {
      this.loadConversations(); // Reload conversations
    },
    logOut() {
      this.$router.push({ path: "/" });
    },
    newItem() {
      console.log("New item triggered");
    }
  },
  mounted() {
    this.username = localStorage.getItem("name") || "Guest";
    this.loadConversations();
  }
};
</script>

<template>
  <div>
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Home Page</h1>
      <p class="username-display">Welcome, {{ username }}!</p>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">Refresh</button>
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="logOut">Log Out</button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newItem">New</button>
        </div>
      </div>
    </div>

    <ErrorMsg v-if="errormsg" :msg="errormsg" />

    <div>
      <h3>My Conversations</h3>

      <p v-if="loading">Loading...</p>

      <div v-else-if="conversations.length === 0">
        <p>No conversations found.</p>
      </div>

      <ul v-else>
        <li v-for="conv in conversations" :key="conv.id">
          <strong @click="viewConversation(conv.id)" style="cursor: pointer; color: #007bff;">
            {{ conv.name }}
          </strong> 
          <div v-if="conv.lastMessage">
            Last message: {{ conv.lastMessage.content }} 
            at {{ new Date(conv.lastMessage.timestamp).toLocaleString() }}
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<style>
.username-display {
  font-size: 16px;
  color: #555;
  margin-top: -10px;
  margin-bottom: 20px;
}
</style>
