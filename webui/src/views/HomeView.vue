<script>
export default {
  data() {
    localStorage.removeItem("recipientId")
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
    viewConversation(conversationId, conversationName) {   
      localStorage.setItem("conversationName", conversationName);
      this.$router.push({
        path: `/conversations/${conversationId}`
      });
    },
    truncateText(text, length = 50, clamp = '...') {
      if (!text || text.length <= length) {
        return text;
      }
      const lastSpaceIndex = text.substring(0, length).lastIndexOf(' ');
      if (lastSpaceIndex === -1) {
        return text.substring(0, length) + clamp;
      }
      return text.substring(0, lastSpaceIndex) + clamp;
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
      <h1 class="h2">My Conversations</h1>
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
      <p v-if="loading">Loading...</p>
      <div v-else-if="conversations.length === 0">
        <p>No conversations found.</p>
      </div>
      <div v-else class="conversations-container">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="conversation-block"
          @click="viewConversation(conv.id, conv.name)"
        >
          <h4>{{ conv.name }}</h4>
          <p v-if="conv.lastMessage">
            Last message: {{ truncateText(conv.lastMessage.content) }} at {{ new Date(conv.lastMessage.timestamp).toLocaleString() }}
          </p>
        </div>
      </div>
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

.conversations-container {
  display: flex;
  flex-direction: column;
}

.conversation-block {
  background-color: #f0f0f0; /* Light grey background */
  padding: 15px;
  margin-bottom: 10px;
  cursor: pointer;
  border-radius: 5px;
}

.conversation-block h4 {
  margin-top: 0;
}

.conversation-block p {
  margin-bottom: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

@media (max-width: 600px) {
  .conversation-block p {
    -webkit-line-clamp: 3;
    line-clamp: 3;
  }
}
</style>