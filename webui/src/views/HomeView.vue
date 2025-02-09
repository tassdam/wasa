<script>
export default {
  data() {
    localStorage.removeItem("recipientId")
    return {
      username: "", 
      errormsg: null,
      loading: false,
      conversations: [],
      pollIntervalId: null
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
        const response = await this.$axios.get("/conversations", {
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
      this.loadConversations(); 
    },
    logOut() {
      this.$router.push({ path: "/" });
    },
    newGroup() {
      this.$router.push({ path: "/new-group" });
    }
  },
  mounted() {
    this.username = localStorage.getItem("name") || "Guest";
    this.loadConversations();
    this.pollIntervalId = setInterval(() => {
      this.loadConversations();
    }, 1000);
  },
  unmounted() {
    clearInterval(this.pollIntervalId);
  }
};
</script>

<template>
  <div>
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">{{ username }}, here is your conversations</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">Refresh</button>
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="logOut">Log Out</button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newGroup">New group</button>
        </div>
      </div>
    </div>
    <ErrorMsg v-if="errormsg" :msg="errormsg" />
    <div>
      <div v-if="conversations.length === 0">
        <p>No conversations found.</p>
      </div>
      <div v-else class="conversations-container">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="conversation-block"
          @click="viewConversation(conv.id, conv.name)"
        >
          <div class="conversation-photo">
            <img
              v-if="conv.conversationPhoto.String"
              :src="'data:image/png;base64,' + conv.conversationPhoto.String"
              alt="Profile Picture"
              class="profile-picture"
            />
          </div>
          <div class="conversation-details">
            <h4>{{ conv.name }}</h4>
            <p v-if="conv.lastMessage" class="last-message">
              Last message by {{ conv.lastMessage.senderName }}: <img v-if="conv.lastMessage.attachment"
             :src="'data:image/*;base64,' + conv.lastMessage.attachment"
             class="attachment-thumbnail"
             alt="Attachment"> {{ truncateText(conv.lastMessage.content) }} at {{ new Date(conv.lastMessage.timestamp).toLocaleString() }}
            </p>
          </div>
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

.conversation-block {
  display: flex; /* Enable flexbox */
  align-items: center; /* Center items vertically */
  gap: 15px; /* Space between photo and text */
}

.conversation-photo {
  flex-shrink: 0; /* Prevent photo from shrinking */
  width: 75px; /* Match profile picture width */
  height: 75px; /* Match profile picture height */
}

.profile-picture {
  width: 75px;
  height: 75px;
  object-fit: cover;
  border-radius: 50%;
}

.last-message {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0;
}

.attachment-thumbnail {
  width: 20px;
  height: 20px;
  object-fit: cover;
  border-radius: 3px;
  flex-shrink: 0;
}

@media (max-width: 600px) {
  .conversation-block p {
    -webkit-line-clamp: 3;
    line-clamp: 3;
  }
}
</style>