<template>
  <div class="chat-container">

    <div class="chat-header">
      <h3>{{ userName }}</h3>
    </div>

    <div class="chat-messages" ref="chatMessages">
      <p v-if="messages.length === 0">No messages yet...</p>
      <div v-for="message in messages" :key="message.id" class="message">
        <p>
          <strong>{{ message.senderName || "Unknown Sender" }}</strong>: 
          {{ message.content }}
        </p>
        <small>{{ formatTimestamp(message.timestamp) }}</small>
      </div>
    </div>

    <div class="chat-input">
      <button class="attach-button">Attach Image or GIF</button>
      <input
        v-model="message"
        class="message-input"
        type="text"
        placeholder="Type a message..."
        @input="toggleSendButton"
      />
      <button
        v-if="message.trim()"
        class="send-button"
        @click="sendMessage"
      >
        Send
      </button>
    </div>
  </div>
</template>

<script>
import axios from "../services/axios";

export default {
  name: "ChatView",
  data() {
    return {
      message: "",
      messages: [], 
      userName: localStorage.getItem("recipientName") || "Unknown User", 
      conversationId: this.$route.params.uuid, 
    };
  },
  methods: {
    async sendMessage() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        const response = await axios.post(
          `/conversations/${this.conversationId}/messages`,
          { content: this.message },
          {
            headers: {
              Authorization: `Bearer ${token}`, 
            },
          }
        );
        this.message = ""; 
        this.fetchMessages(); 
      } catch (error) {
        console.error("Failed to send message:", error);
      }
    },

    async fetchMessages() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        const response = await axios.get(
          `/conversations/${this.conversationId}`,
          {
            headers: {
              Authorization: `Bearer ${token}`, 
            },
          }
        );
        this.messages = response.data.messages.map((message) => ({
          ...message,
          senderName: this.getSenderName(message.senderId),
        }));
        this.$nextTick(() => {
          this.scrollToBottom();
        });
      } catch (error) {
        console.error("Failed to fetch messages:", error);
        alert("Failed to load messages. Please try again later.");
      }
    },

    getSenderName(userId) {
      if (userId === localStorage.getItem("token")) {
        return "You";
      }
      return localStorage.getItem("recipientName") || "Unknown User";
    },

    formatTimestamp(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleString();
    },

    scrollToBottom() {
      const chatMessages = this.$refs.chatMessages;
      if (chatMessages) {
        chatMessages.scrollTop = chatMessages.scrollHeight;
      }
    },
  },
  mounted() {
    this.fetchMessages();
  },
};
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 92vh; 
  overflow: hidden;
}

.chat-header {
  padding: 15px;
  font-size: 20px;
  font-weight: bold;
  text-align: left;
}

.chat-messages {
  flex: 1; 
  overflow-y: auto; 
  padding: 20px;
  border-top: 1px solid #ccc;
  border-bottom: 1px solid #ccc;
}

.message {
  margin-bottom: 10px;
}

.chat-input {
  display: flex;
  align-items: center;
  padding: 10px;
  background-color: white;
  border-top: 1px solid #ccc;
  position: sticky; 
  bottom: 0;
}

.attach-button {
  background-color: #25d366; 
  color: white;
  border: none;
  padding: 10px;
  border-radius: 5px;
  cursor: pointer;
  margin-right: 10px;
}

.attach-button:hover {
  background-color: #20b358;
}

.message-input {
  flex: 1; 
  padding: 12px;
  border: 1px solid #ccc;
  border-radius: 5px;
  font-size: 16px;
}

.send-button {
  background-color: #128c7e; 
  color: white;
  border: none;
  padding: 12px 20px;
  border-radius: 5px;
  margin-left: 10px;
  cursor: pointer;
}

.send-button:hover {
  background-color: #0f7c6a;
}
</style>
