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
          <p v-if="message.content.startsWith('<strong>Forwarded from')"
            v-html="message.content">
          </p>
          <p v-else>
            <strong>{{ message.senderId === userToken ? 'You' : (message.senderName || 'Unknown Sender') }}:</strong>
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
            class="action-button heart-button"
            :style="getActionButtonStyle(message)"
            :class="{ 'has-reacted': (message.reactingUserIDs || []).includes(userToken) }"
            @click.stop="toggleReaction(message)"
          >
            ❤️
          </button>
          <button 
            class="action-button forward-button"
            :style="getForwardButtonStyle(message)"
            @click.stop="showForwardOptions(message.id)"
          >
            →
          </button>
          <button 
            v-if="message.senderId === userToken"
            class="action-button forward-button"
            :style="getForwardButtonStyle(message)"
            @click.stop="showForwardOptions(message.id)"
          >
            →
          </button>
          <button
            v-if="message.senderId === userToken"
            class="action-button delete-button"
            :style="getActionButtonStyle(message)"
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
              <option value="" disabled>Select conversation</option>
              <option
                v-for="conv in messageOptions[message.id].forwardConversations"
                :key="conv.id"
                :value="conv.id"
              >
                {{ conv.name }}
              </option>
              <option value="new">New contact</option>
            </select>
            <div v-if="messageOptions[message.id].selectedConversationId === 'new'" class="contact-search">
              <input
                type="text"
                v-model="messageOptions[message.id].contactQuery"
                placeholder="Enter contact name"
                @input="searchContact(message.id)"
              />
              <ul v-if="messageOptions[message.id].contactResults.length > 0" class="contact-results">
                <li
                  v-for="contact in messageOptions[message.id].contactResults"
                  :key="contact.id"
                  @click="selectContact(contact, message.id)"
                  class="contact-result"
                >
                  {{ contact.name }}
                </li>
              </ul>
            </div>
            <div class="forward-buttons-container">
              <button
                class="button-style"
                v-if="messageOptions[message.id].selectedConversationId !== 'new'"
                :disabled="!messageOptions[message.id].selectedConversationId"
                @click.stop="forwardMessage(messageOptions[message.id].selectedConversationId, message.id)"
              >
                Send
              </button>
              <button
                class="button-style"
                v-if="messageOptions[message.id].selectedConversationId === 'new'"
                :disabled="!messageOptions[message.id].selectedContactId"
                @click.stop="forwardToContact(messageOptions[message.id].selectedContactId, message.id)"
              >
                Send
              </button>
              <button class="button-style" @click.stop="closeForwardMenu(message.id)">Cancel</button>
            </div>
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
        <span v-if="selectedFile" class="file-icon">🖼️</span>
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
          selectedConversationId: "",
          contactQuery: "",
          contactResults: [],
          selectedContactId: ""
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
        // Exclude the current conversation
        const conversations = response.data.filter(conv => conv.id !== this.conversationId);
        this.messageOptions[messageId].forwardConversations = conversations;
      } catch (error) {
        console.error("Failed to fetch conversations:", error);
        alert("Failed to fetch conversations. Please try again.");
      }
    },
    async searchContact(messageId) {
      const query = this.messageOptions[messageId].contactQuery;
      if (!query.trim()) {
        this.messageOptions[messageId].contactResults = [];
        return;
      }
      try {
        const token = localStorage.getItem("token");
        const response = await axios.get('/search', {
          params: { username: query },
          headers: { Authorization: `Bearer ${token}` }
        });
        this.messageOptions[messageId].contactResults = response.data;
      } catch (error) {
        console.error("Contact search failed:", error);
        this.messageOptions[messageId].contactResults = [];
      }
    },
    selectContact(contact, messageId) {
      this.messageOptions[messageId].selectedContactId = contact.id;
      this.messageOptions[messageId].contactQuery = contact.name;
      this.messageOptions[messageId].contactResults = [];
    },
    async forwardToContact(selectedContactId, messageId) {
      const token = localStorage.getItem("token");
      if (!token) {
        this.$router.push({ path: "/" });
        return;
      }
      let conversationResponse;
      try {
        conversationResponse = await axios.post(
          `/conversations`,
          { senderId: token, recipientId: selectedContactId },
          { headers: { Authorization: `Bearer ${token}` } }
        );
      } catch (error) {
        console.error("Error creating conversation:", error);
        alert("Failed to create a conversation with the specified contact.");
        return;
      }
      const targetConversationId = conversationResponse.data.conversationId;
      try {
        await axios.post(
          `/conversations/${this.conversationId}/message/${messageId}/forward`,
          { sourceMessageId: messageId, targetConversationId: targetConversationId },
          { headers: { Authorization: `Bearer ${token}` } }
        );
        alert("Message forwarded successfully!");
        this.closeForwardMenu(messageId);
      } catch (error) {
        console.error("Failed to forward message:", error);
        alert("Failed to forward message. Please try again.");
      }
    },
    async forwardMessage(targetConversationId, messageId) {
      const message = this.messages.find(m => m.id === messageId);
      if (!message) return;
      try {
        const token = localStorage.getItem("token");
        const forwarderName = localStorage.getItem("name") || "Unknown";
        await axios.post(
          `/conversations/${this.conversationId}/message/${messageId}/forward`,
          { targetConversationId: targetConversationId, forwarderName: forwarderName },
          { headers: { Authorization: `Bearer ${token}` } }
        );
        alert("Message forwarded successfully!");
        this.closeForwardMenu(messageId);
      } catch (error) {
        console.error("Failed to forward message:", error);
        alert("Failed to forward message. Please try again.");
      }
    },
    getForwardButtonStyle(message) {
      return message.senderId === this.userToken
        ? { left: '-40px', top: 'calc(50% - 20px)' }
        : { right: '-40px', top: 'calc(50% - 20px)' };
    },
    getActionButtonStyle(message) {
      return message.senderId === this.userToken
        ? { left: '-40px', top: 'calc(50% + 5px)' }
        : { right: '-40px', top: 'calc(50% + 5px)' };
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
  position: relative;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
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
.message-content {
  position: relative;
  padding-right: 80px; 
  min-height: 40px;
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

.chat-input {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  padding: 10px;
  gap: 8px;
}
.attach-button {
  background-color: #25d366;
  color: white;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  margin-right: 10px;
  font-size: 14px;
  padding: 10px 15px;
}
.attach-button:hover {
  background-color: #20b358;
}
.message-input {
  flex: 1;
  min-width: 200px;
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

.action-button {
  position: absolute;
  width: 24px;
  height: 24px;
  border: 1px solid #aaa;
  border-radius: 50%;
  background-color: white;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
  opacity: 0.9;
  transition: opacity 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  padding: 0;
}
.action-button:hover {
  opacity: 1;
  background-color: white;
}

/* Forward options menu */
.forward-options {
  position: absolute;
  top: 30px;
  right: 0;
  background-color: #ffffff;
  border-radius: 5px;
  padding: 10px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  z-index: 100;
  width: 250px;
}
.forward-select {
  width: 100%;
  padding: 8px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  margin-bottom: 8px;
  font-size: 14px;
}
.forward-buttons-container {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-top: 10px;
}
.button-style {
  background-color: #128c7e;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.button-style:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.contact-search input {
  width: 100%;
  padding: 6px;
  margin-bottom: 4px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.contact-results {
  list-style: none;
  padding: 0;
  margin: 0 0 6px 0;
  max-height: 100px;
  overflow-y: auto;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.contact-result {
  padding: 4px;
  cursor: pointer;
  border-bottom: 1px solid #eee;
}
.contact-result:hover {
  background-color: #f0f0f0;
}

.file-icon {
  font-size: 18px;
  margin-left: 5px;
}

@media (max-width: 600px) {
  .conversation-block p {
    -webkit-line-clamp: 3;
    line-clamp: 3;
  }
}
</style>
