<script>
export default {
  data() {
    return {
      username: "", // Store the logged-in user's name
      errormsg: null,
      loading: false,
      conversations: [] // Store the list of conversations
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
          this.errormsg = "No token found. Are you sure you are logged in?";
          this.loading = false;
          return;
        }

        // Fetch conversations from the server
        const response = await this.$axios.get("/users/me/conversations", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });

        // Update conversations
        this.conversations = response.data || [];
      } catch (error) {
        console.error("Error loading conversations:", error);
        this.errormsg = "Failed to load conversations. Please try again.";
      } finally {
        this.loading = false;
      }
    },
    viewConversation(conversationId, conversationName) {
      // Redirect to the conversation page
      
      this.$router.push({
        path: `/conversations/${conversationId}`
      });
    },
    refresh() {
      console.log("Refresh triggered");
      this.loadConversations(); // Reload conversations
    },
    exportList() {
      console.log("Export triggered");
    },
    newItem() {
      console.log("New item triggered");
    }
  },
  mounted() {
    // Retrieve the username and load conversations on mount
    this.username = localStorage.getItem("name") || "Guest";
    this.loadConversations();
  }
};
</script>

<template>
  <div>
    <!-- Top bar -->
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Home Page</h1>
      <p class="username-display">Welcome, {{ username }}!</p>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">Refresh</button>
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="exportList">Export</button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newItem">New</button>
        </div>
      </div>
    </div>

    <!-- Error Message -->
    <ErrorMsg v-if="errormsg" :msg="errormsg" />

    <!-- Conversations List -->
    <div>
      <h3>My Conversations</h3>

      <!-- Loading Spinner -->
      <p v-if="loading">Loading...</p>

      <!-- No Conversations Message -->
      <div v-else-if="conversations.length === 0">
        <p>No conversations found.</p>
      </div>

      <!-- Conversations -->
      <ul v-else>
        <li v-for="conv in conversations" :key="conv.id">
          <strong @click="viewConversation(conv.id, conv.name)" style="cursor: pointer; color: #007bff;">
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
/* Add styling for the username display */
.username-display {
  font-size: 16px;
  color: #555;
  margin-top: -10px;
  margin-bottom: 20px;
}
</style>
