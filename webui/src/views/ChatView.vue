<template>
  <div class="chat-container">
    <div class="chat-header">
      <h3>{{ convName }}</h3>
    </div>
    <div class="chat-messages" ref="chatMessages">
      <p v-if="messages.length === 0">No messages yet...</p>
      <div v-for="message in messages" :key="message.id" class="message" :class="message.senderId === userToken ? 'self' : 'other'">
        <div class="message-content" @click.stop>
          <p>
            <strong v-if="message.forwardedMessageId">
              Forwarded from {{ message.senderName || 'Unknown Sender' }}:
            </strong>
            <strong v-else>
              {{ message.senderId === userToken ? 'You' : (message.senderName || 'Unknown Sender') }}:
            </strong>
            {{ message.content }}
          </p>
          <small>{{ formatTimestamp(message.timestamp) }}</small>
          <div v-if="message.reactionCount > 0" class="reaction-count">
            ❤️ × {{ message.reactionCount }}
          </div>
          <button 
            v-if="message.senderId !== userToken"
            class="heart-button"
            :class="{ 'has-reacted': (message.reactingUserIDs || []).includes(userToken) }"
            @click.stop="toggleReaction(message)"
          >
            ❤️
          </button>
          <button class="forward-button" @click.stop="showForwardOptions(message.id)">→</button>
          <button v-if="message.senderId === userToken" class="delete-button" @click.stop="deleteMessage(message)">✖</button>
          <div v-if="messageOptions[message.id]?.showForwardMenu" class="forward-options" @click.stop>
            <label for="forward-select">Forward to:</label>
            <select
              id="forward-select"
              class="forward-select"
              v-model="messageOptions[message.id].selectedConversationId"
            >
              <option
                v-for="conv in messageOptions[message.id].forwardConversations"
                :key="conv.id"
                :value="conv.id"
              >
                {{ conv.name }}
              </option>
            </select>
            <button
              class="button-style"
              :disabled="!messageOptions[message.id].selectedConversationId"
              @click.stop="forwardMessage(messageOptions[message.id].selectedConversationId, message.id)"
            >
              Send
            </button>
            <button class="button-style" @click.stop="closeForwardMenu(message.id)">Cancel</button>
            <div v-if="messageOptions[message.id].forwardConversations.length === 0">
              No conversation found.
            </div>
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
          `/conversations/${this.conversationId}/message`,
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
        // Ensure reactingUserIDs exists as an array
        this.messages = (response.data.messages || []).map(msg => ({
          ...msg,
          reactingUserIDs: msg.reactingUserIDs || []
        }));
        this.$nextTick(() => {
          this.scrollToBottom();
        });
      } catch (error) {
        console.error("Failed to fetch messages:", error);
        alert("Failed to load messages. Please try again later.");
      }
    },
    async toggleReaction(message) {
      try {
        const token = localStorage.getItem("token");
        if (!token || message.senderId === this.userToken) return;

        const hasReacted = (message.reactingUserIDs || []).includes(this.userToken);
        
        if (hasReacted) {
          await axios.delete(
            `/conversations/${this.conversationId}/messages/${message.id}/comment`,
            { headers: { Authorization: `Bearer ${token}` } }
          );
          const userIndex = message.reactingUserIDs.indexOf(this.userToken);
          if (userIndex > -1) {
            message.reactingUserIDs.splice(userIndex, 1);
            message.reactionCount = Math.max(0, message.reactionCount - 1);
          }
        } else {
          await axios.post(
            `/conversations/${this.conversationId}/messages/${message.id}/comment`,
            { headers: { Authorization: `Bearer ${token}` } }
          );
          // Update local state
          if (!message.reactingUserIDs) {
            message.reactingUserIDs = [];
          }
          message.reactingUserIDs.push(this.userToken);
          message.reactionCount = (message.reactionCount || 0) + 1;
        }
      } catch (error) {
        console.error("Failed to toggle reaction:", error);
        alert("Failed to update reaction. Please try again.");
      }
    },
    async deleteMessage(message) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        await axios.delete(
          `/conversations/${this.conversationId}/messages/${message.id}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
        this.messages = this.messages.filter(m => m.id !== message.id);
        this.$nextTick(() => {
          this.scrollToBottom();
        });
      } catch (error) {
        console.error("Failed to delete message:", error);
        alert("Failed to delete message. Please try again later.");
      }
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
    showForwardOptions(messageId) {
      this.closeAllMenus();
      if (!this.messageOptions[messageId]) {
        this.messageOptions[messageId] = {
          showForwardMenu: true,
          forwardConversations: [],
          selectedConversationId: null,
        };
        this.fetchForwardConversations(messageId);
      } else {
        this.messageOptions[messageId].showForwardMenu = !this.messageOptions[messageId].showForwardMenu;
      }
    },
    closeForwardMenu(messageId) {
      if (this.messageOptions[messageId]) {
        this.messageOptions[messageId].showForwardMenu = false;
      }
    },
    closeAllMenus() {
      for (const id in this.messageOptions) {
        this.messageOptions[id].showForwardMenu = false;
      }
    },
    handleOutsideClick(event) {
      const messageContent = this.$el.querySelector('.message-content');
      if (
        messageContent &&
        !messageContent.contains(event.target)
      ) {
        this.closeAllMenus();
      }
    },
    async fetchForwardConversations(messageId) {
      try {
        const token = localStorage.getItem("token");
        const response = await axios.get('/conversations', {
          headers: { Authorization: `Bearer ${token}` },
        });
        const conversations = response.data.filter(conv => conv.id !== this.conversationId);
        this.messageOptions[messageId].forwardConversations = conversations;
      } catch (error) {
        console.error("Failed to fetch conversations:", error);
        alert("Failed to fetch conversations. Please try again.");
      }
    },
    async forwardMessage(targetConversationId, messageId) {
      const message = this.messages.find(m => m.id === messageId);
      if (!message) return;
      try {
        const token = localStorage.getItem("token");
        await axios.post(
          `/conversations/${this.conversationId}/messages/${messageId}/forward`,
          { 
            sourceMessageId: message.id,
            targetConversationId: targetConversationId 
          },
          {
            headers: { Authorization: `Bearer ${token}` },
          }
        );
        alert("Message forwarded successfully!");
        this.closeForwardMenu(messageId);
      } catch (error) {
        console.error("Failed to forward message:", error);
        alert("Failed to forward message. Please try again.");
      }
    },
  },
  mounted() {
    this.fetchMessages();
    document.addEventListener('click', this.handleOutsideClick);
  },
  beforeDestroy() {
    document.removeEventListener('click', this.handleOutsideClick);
  },
};
</script>

<style scoped>
.message-content {
  position: relative;
  box-sizing: border-box;
}

.heart-button, .forward-button, .delete-button {
  background-color: transparent;
  border: none;
  position: absolute;
  top: 1px;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 5px;
}

.heart-button {
  right: 40px;
  color: #666;
}

.heart-button.has-reacted {
  color: #ff3860;
}

.forward-button {
  right: 20px;
}

.delete-button {
  right: 1px;
}

.message-content:hover .heart-button,
.message-content:hover .forward-button,
.message-content:hover .delete-button {
  display: block;
}

.reaction-count {
  margin-top: 4px;
  font-size: 0.9em;
  color: #666;
  display: flex;
  align-items: center;
  gap: 4px;
}

.forward-options {
  position: absolute;
  top: 30px;
  right: 0;
  background-color: #ffffff;
  border-radius: 5px;
  padding: 10px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  z-index: 100;
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
  box-sizing: border-box;
  position: relative;
}

.message.self {
  margin-left: auto;
  background-color: #d1e7dd;
  padding: 10px;
  border-radius: 10px;
}

.message.other {
  background-color: #e0f2f1;
  padding: 10px;
  border-radius: 10px;
}

.message p {
  margin: 0;
  color: #333;
  word-wrap: break-word;
  word-break: break-word;
  white-space: pre-wrap;
}

.message small {
  margin-top: 5px;
  color: #666;
  display: block;
  font-size: 0.8em;
}

.message.self small {
  text-align: right;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 92vh;
  overflow: hidden;
}

.chat-header {
  padding: 15px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
}

.chat-input {
  display: flex;
  align-items: center;
  padding: 10px;
  background-color: white;
  border-top: 1px solid #dee2e6;
  position: sticky;
  bottom: 0;
}

.attach-button {
  background-color: #25d366;
  color: white;
  border: none;
  padding: 10px 15px;
  border-radius: 20px;
  cursor: pointer;
  margin-right: 10px;
  font-size: 14px;
}

.attach-button:hover {
  background-color: #20b358;
}

.message-input {
  flex: 1;
  padding: 12px;
  border: 1px solid #dee2e6;
  border-radius: 20px;
  font-size: 14px;
  outline: none;
}

.send-button {
  background-color: #128c7e;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 20px;
  margin-left: 10px;
  cursor: pointer;
  font-size: 14px;
}

.send-button:hover {
  background-color: #0f7c6a;
}

.forward-select {
  width: 200px;
  padding: 8px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  margin-bottom: 8px;
  font-size: 14px;
}

.button-style {
  background-color: #128c7e;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  font-size: 14px;
}

.button-style:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.button-style:hover:not(:disabled) {
  background-color: #0f7c6a;
}
</style>