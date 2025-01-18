<template>
  <div class="chat-container">
    <div class="chat-header">
      <h3>{{ convName }}</h3>
    </div>
    <div class="chat-messages" ref="chatMessages">
      <p v-if="messages.length === 0">No messages yet...</p>
      <div v-for="message in messages" :key="message.id" class="message" :class="message.senderId === userToken ? 'self' : 'other'">
        <div class="message-content" @mouseover="hoverMessage(message.id)" @mouseout="unhoverMessage(message.id)">
          <p>
            <strong>{{ message.senderId === userToken ? 'You' : (message.senderName || 'Unknown Sender') }}</strong>:
            {{ message.content }}
          </p>
          <small>{{ formatTimestamp(message.timestamp) }}</small>
          <button class="options-button" @click="toggleOptions(message.id)" v-if="hoveredMessages[message.id]">OptionsMenu</button>
          <div class="options-menu" v-if="messageOptions[message.id]">
            <button @click="forwardMessage(message)">Forward</button>
            <button @click="commentOnMessage(message)">Comment</button>
            <button v-if="message.senderId === userToken" @click="deleteMessage(message)">Delete</button>
          </div>
        </div>
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
      userToken: localStorage.getItem("token"),
      convName: localStorage.getItem("conversationName") || "Unknown User", 
      conversationId: this.$route.params.uuid, 
      messageOptions: {},
      hoveredMessages: {}
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
        this.messages = response.data.messages;
        this.$nextTick(() => {
          this.scrollToBottom();
        });
      } catch (error) {
        console.error("Failed to fetch messages:", error);
        alert("Failed to load messages. Please try again later.");
      }
    },

    async deleteMessage(message) {
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

    toggleOptions(messageId) {
      if (this.messageOptions[messageId] === undefined) {
        this.messageOptions[messageId] = true;
      } else {
        this.messageOptions[messageId] = !this.messageOptions[messageId];
      }
    },
    
    hoverMessage(messageId) {
      this.hoveredMessages[messageId] = true;
    },
    
    unhoverMessage(messageId) {
      this.hoveredMessages[messageId] = false;
    },

    closeOptions(event) {
      if (!this.$el.contains(event.target)) {
        this.messageOptions = {};
      }
    },

    forwardMessage(message) {
      // Implement forward message logic
    },

    commentOnMessage(message) {
      // Implement comment on message logic
    },
  },
  mounted() {
    this.fetchMessages();
    document.addEventListener('click', this.closeOptions);
  },
};
</script>

<style scoped>

.message-content {
  position: relative;
}

.options-button {
  display: none;
  position: absolute;
  top: 5px;
  right: 5px;
}

.message:hover .options-button {
  display: block;
}

.options-menu {
  position: absolute;
  top: 30px;
  right: 0;
  background-color: white;
  border: 1px solid #ccc;
  border-radius: 5px;
  padding: 5px;
}

.chat-messages {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  border-top: 1px solid #ccc;
  border-bottom: 1px solid #ccc;
}

.message {
  max-width: 70%;
  margin-bottom: 10px;
  display: flex;
  flex-direction: column;
}

.message.self {
  margin-left: auto;
  background-color: #d1e7dd;
  padding: 10px;
  border-radius: 5px;
}

.message.other {
  background-color: #e0f2f1;
  padding: 10px;
  border-radius: 5px;
}

.message p {
  margin: 0;
  color: #333;
}

.message small {
  margin-top: 5px;
  color: #666;
}

.message.self small {
  align-self: flex-end;
}

.message.other small {
  align-self: flex-start;
}

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
