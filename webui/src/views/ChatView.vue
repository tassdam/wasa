<template>
  <div class="chat-container">
    <div class="chat-header">
      <h3>{{ convName }}</h3>
    </div>
    <div class="chat-messages" ref="chatMessages">
      <p v-if="messages.length === 0">No messages yet...</p>
      <div
        v-for="message in messages"
        :key="message.id"
        class="message"
        :class="message.senderId === userToken ? 'self' : 'other'"
      >
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
          <div v-if="message.attachment" class="attachment-container">
            <img 
              :src="'data:image/png;base64,' + message.attachment" 
              alt="Attachment"
              class="attachment-image"
            />
          </div>
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
          <button
            v-if="message.senderId === userToken"
            class="delete-button"
            @click.stop="deleteMessage(message)"
          >
            ✖
          </button>
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
      <input 
        type="file" 
        ref="fileInput" 
        style="display: none" 
        accept="image/*, .gif"
        @change="handleFileSelect"
      >
      <button class="attach-button" @click="triggerFileInput">
        Attach Image or GIF
        <span v-if="selectedFile" class="file-name">{{ selectedFile.name }}</span>
      </button>
      <input
        v-model="message"
        class="message-input"
        type="text"
        placeholder="Type a message..."
        @input="toggleSendButton"
      />
      <button
        v-if="message.trim() || selectedFile"
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
      selectedFile: null,
      pollIntervalId: null,
      firstLoad: true  
    };
  },
  methods: {
    triggerFileInput() {
      this.$refs.fileInput.click();
    },
    handleFileSelect(event) {
      this.selectedFile = event.target.files[0];
    },
    async sendMessage() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        const formData = new FormData();
        formData.append("content", this.message);
        if (this.selectedFile) {
          formData.append("attachment", this.selectedFile);
        }
        await axios.post(
          `/conversations/${this.conversationId}/message`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );
        
        this.message = "";
        this.selectedFile = null;
        this.$refs.fileInput.value = "";
        
        await this.fetchMessages(); 
        this.$nextTick(() => {
          this.forceScrollToBottom();
        });
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
              Authorization: `Bearer ${token}`
            }
          }
        );
        this.messages = (response.data.messages || []).map(msg => ({
          ...msg,
          reactingUserIDs: msg.reactingUserIDs || []
        }));
        
        this.$nextTick(() => {
          if (this.firstLoad) {
            this.forceScrollToBottom();
            this.firstLoad = false;
          }
        });
      } catch (error) {
        console.error("Failed to fetch messages:", error);
        alert("Failed to load messages. Please try again later.");
      }
    },
    forceScrollToBottom() {
      const chat = this.$refs.chatMessages;
      if (chat) {
        chat.scrollTop = chat.scrollHeight;
      }
    },
    async toggleReaction(message) {
      try {
        const token = localStorage.getItem("token");
        if (!token || message.senderId === this.userToken) return;
        const hasReacted = (message.reactingUserIDs || []).includes(token);
        if (hasReacted) {
          await axios.delete(
            `/conversations/${this.conversationId}/message/${message.id}/comment`,
            { headers: { Authorization: `Bearer ${token}` } }
          );
          const userIndex = message.reactingUserIDs.indexOf(this.userToken);
          if (userIndex > -1) {
            message.reactingUserIDs.splice(userIndex, 1);
            message.reactionCount = Math.max(0, message.reactionCount - 1);
          }
        } else {
          await axios.post(
            `/conversations/${this.conversationId}/message/${message.id}/comment`,
            {},
            { headers: { Authorization: `Bearer ${token}` } }
          );
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
          `/conversations/${this.conversationId}/message/${message.id}`,
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );
        this.messages = this.messages.filter(m => m.id !== message.id);
      } catch (error) {
        console.error("Failed to delete message:", error);
        alert("Failed to delete message. Please try again later.");
      }
    },
    formatTimestamp(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleString();
    },
    showForwardOptions(messageId) {
      this.closeAllMenus();
      if (!this.messageOptions[messageId]) {
        this.messageOptions[messageId] = {
          showForwardMenu: true,
          forwardConversations: [],
          selectedConversationId: null
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
      if (messageContent && !messageContent.contains(event.target)) {
        this.closeAllMenus();
      }
    },
    async fetchForwardConversations(messageId) {
      try {
        const token = localStorage.getItem("token");
        const response = await axios.get('/conversations', {
          headers: { Authorization: `Bearer ${token}` }
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
          `/conversations/${this.conversationId}/message/${messageId}/forward`,
          { 
            sourceMessageId: message.id,
            targetConversationId: targetConversationId 
          },
          {
            headers: { Authorization: `Bearer ${token}` }
          }
        );
        alert("Message forwarded successfully!");
        this.closeForwardMenu(messageId);
      } catch (error) {
        console.error("Failed to forward message:", error);
        alert("Failed to forward message. Please try again.");
      }
    }
  },
  mounted() {
    this.fetchMessages();
    this.pollIntervalId = setInterval(() => {
      this.fetchMessages();
    }, 100);
    document.addEventListener("click", this.handleOutsideClick);
  },
  beforeUnmount() {
    document.removeEventListener("click", this.handleOutsideClick);
    clearInterval(this.pollIntervalId);
  }
};
</script>

<style scoped>
/* Your existing styles remain unchanged */
.attachment-container {
  margin-top: 8px;
  max-width: 300px;
}
.attachment-image {
  max-width: 100%;
  border-radius: 8px;
  border: 1px solid #ddd;
  margin-top: 4px;
}
.attachment-name {
  display: block;
  color: #666;
  font-size: 0.75rem;
  margin-top: 4px;
  word-break: break-all;
}
.file-input-container {
  position: relative;
  margin-right: 10px;
}
.file-name {
  display: block;
  font-size: 0.75rem;
  color: #666;
  margin-top: 4px;
  max-width: 120px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.chat-input {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  padding: 10px;
  gap: 8px;
}
.attach-button {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 15px;
}
.message-input {
  flex: 1;
  min-width: 200px;
}
.message-content {
  position: relative;
  box-sizing: border-box;
  padding-right: 80px;
  min-height: 40px;
}
.heart-button, .forward-button, .delete-button {
  position: absolute;
  top: 5px;
  background-color: rgba(255, 255, 255, 0.9);
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 2px 6px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  opacity: 0.8;
  transition: opacity 0.2s;
  display: block;
}
.heart-button:hover, 
.forward-button:hover, 
.delete-button:hover {
  opacity: 1;
  background-color: white;
}
.delete-button {
  right: 5px;
}
.forward-button {
  right: 30px;
}
.heart-button {
  right: 55px;
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
  margin-right: -60px;
  padding-right: 10px;
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
.attach-button {
  background-color: #25d366;
  color: white;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  margin-right: 10px;
  font-size: 14px;
}
.attach-button:hover {
  background-color: #20b358;
}
.message-input {
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
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  font-size: 14px;
}
</style>
