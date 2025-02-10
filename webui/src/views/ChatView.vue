<template>
  <div class="chat-container">
    <div class="chat-header">
      <div class="chat-photo" v-if="conversationPhoto">
        <img :src="'data:image/jpeg;base64,' + conversationPhoto" alt="Chat Thumbnail" />
      </div>
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
        <div v-if="conversationType==='group' && message.senderId !== userToken" class="sender-thumbnail">
          <img :src="'data:image/jpeg;base64,' + message.senderPhoto" alt="Sender Photo" />
        </div>
        <div class="message-content" @click.stop>
          <p v-if="message.content.startsWith('<strong>Forwarded from')" v-html="message.content"></p>
          <p v-else>
            <strong>
              {{ message.senderId === userToken ? 'You' : (message.senderName || 'Unknown Sender') }}:
            </strong>
            {{ message.content }}
          </p>
          <div v-if="message.attachment" class="attachment-container">
            <img :src="'data:image/jpeg;base64,' + message.attachment" alt="Attachment" class="attachment-image" />
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
            class="action-button delete-button"
            :style="getActionButtonStyle(message)"
            @click.stop="deleteMessage(message)"
          >
            ✖
          </button>
          <div v-if="messageOptions[message.id]?.showForwardMenu" class="forward-options" @click.stop>
            <select id="forward-select" class="forward-select" v-model="messageOptions[message.id].selectedConversationId">
              <option value="" disabled>Select conversation</option>
              <option v-for="conv in messageOptions[message.id].forwardConversations" :key="conv.id" :value="conv.id">
                {{ conv.name }}
              </option>
              <option value="new">New contact</option>
            </select>
            <div v-if="messageOptions[message.id].selectedConversationId === 'new'" class="contact-search">
              <input type="text" v-model="messageOptions[message.id].contactQuery" placeholder="Enter contact name" @input="searchContact(message.id)" />
              <ul v-if="messageOptions[message.id].contactResults.length > 0" class="contact-results">
                <li v-for="contact in messageOptions[message.id].contactResults" :key="contact.id" @click="selectContact(contact, message.id)" class="contact-result">
                  {{ contact.name }}
                </li>
              </ul>
            </div>
            <div class="forward-buttons-container">
              <button class="button-style" v-if="messageOptions[message.id].selectedConversationId !== 'new'" :disabled="!messageOptions[message.id].selectedConversationId" @click.stop="forwardMessage(messageOptions[message.id].selectedConversationId, message.id)">
                Send
              </button>
              <button class="button-style" v-if="messageOptions[message.id].selectedConversationId === 'new'" :disabled="!messageOptions[message.id].selectedContactId" @click.stop="forwardToContact(messageOptions[message.id].selectedContactId, message.id)">
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
      <input type="file" ref="fileInput" style="display: none" accept="image/*, .gif" @change="handleFileSelect" />
      <button class="attach-button" @click="triggerFileInput">
        Attach Image or GIF
        <span v-if="selectedFile" class="file-icon">🖼️</span>
      </button>
      <input v-model="message" class="message-input" type="text" placeholder="Type a message..." @input="toggleSendButton" />
      <button v-if="message.trim() || selectedFile" class="send-button" @click="sendMessage">
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
      conversationPhoto: null,
      conversationType: null,
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
      await axios.post(`/conversations/${this.conversationId}/message`, formData, { headers: { Authorization: `Bearer ${token}` } });
      this.message = "";
      this.selectedFile = null;
      this.$refs.fileInput.value = "";
      await this.fetchMessages();
      this.$nextTick(() => {
        this.forceScrollToBottom();
      });
    },
    async fetchMessages() {
      const token = localStorage.getItem("token");
      if (!token) {
        this.$router.push({ path: "/" });
        return;
      }
      const response = await axios.get(`/conversations/${this.conversationId}`, { headers: { Authorization: `Bearer ${token}` } });
      this.messages = (response.data.messages || []).map(msg => ({ ...msg, reactingUserIDs: msg.reactingUserIDs || [] }));
      if (response.data.name) {
        this.convName = response.data.name;
      }
      if (response.data.conversationPhoto && response.data.conversationPhoto.String) {
        this.conversationPhoto = response.data.conversationPhoto.String;
      } else {
        this.conversationPhoto = null;
      }
      if (response.data.type) {
        this.conversationType = response.data.type;
      } else {
        this.conversationType = "direct";
      }
      this.$nextTick(() => {
        if (this.firstLoad) {
          this.forceScrollToBottom();
          this.firstLoad = false;
        }
      });
    },
    forceScrollToBottom() {
      const chat = this.$refs.chatMessages;
      if (chat) {
        chat.scrollTop = chat.scrollHeight;
      }
    },
    async toggleReaction(message) {
      const token = localStorage.getItem("token");
      if (!token || message.senderId === this.userToken) return;
      const hasReacted = (message.reactingUserIDs || []).includes(token);
      if (hasReacted) {
        await axios.delete(`/conversations/${this.conversationId}/message/${message.id}/comment`, { headers: { Authorization: `Bearer ${token}` } });
        const userIndex = message.reactingUserIDs.indexOf(this.userToken);
        if (userIndex > -1) {
          message.reactingUserIDs.splice(userIndex, 1);
          message.reactionCount = Math.max(0, message.reactionCount - 1);
        }
      } else {
        await axios.post(`/conversations/${this.conversationId}/message/${message.id}/comment`, {}, { headers: { Authorization: `Bearer ${token}` } });
        if (!message.reactingUserIDs) {
          message.reactingUserIDs = [];
        }
        message.reactingUserIDs.push(this.userToken);
        message.reactionCount = (message.reactionCount || 0) + 1;
      }
    },
    async deleteMessage(message) {
      const token = localStorage.getItem("token");
      if (!token) {
        this.$router.push({ path: "/" });
        return;
      }
      await axios.delete(`/conversations/${this.conversationId}/message/${message.id}`, { headers: { Authorization: `Bearer ${token}` } });
      this.messages = this.messages.filter(m => m.id !== message.id);
    },
    formatTimestamp(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleString();
    },
    showForwardOptions(messageId) {
      this.closeAllMenus();
      if (!this.messageOptions[messageId]) {
        this.messageOptions[messageId] = { showForwardMenu: true, forwardConversations: [], selectedConversationId: "", contactQuery: "", contactResults: [], selectedContactId: "" };
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
      const token = localStorage.getItem("token");
      const response = await axios.get('/conversations', { headers: { Authorization: `Bearer ${token}` } });
      const conversations = response.data.filter(conv => conv.id !== this.conversationId);
      this.messageOptions[messageId].forwardConversations = conversations;
    },
    async searchContact(messageId) {
      const query = this.messageOptions[messageId].contactQuery;
      if (!query.trim()) {
        this.messageOptions[messageId].contactResults = [];
        return;
      }
      const token = localStorage.getItem("token");
      const response = await axios.get('/search', { params: { username: query }, headers: { Authorization: `Bearer ${token}` } });
      this.messageOptions[messageId].contactResults = response.data;
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
      conversationResponse = await axios.post(`/conversations`, { senderId: token, recipientId: selectedContactId }, { headers: { Authorization: `Bearer ${token}` } });
      const targetConversationId = conversationResponse.data.conversationId;
      await axios.post(`/conversations/${this.conversationId}/message/${messageId}/forward`, { sourceMessageId: messageId, targetConversationId: targetConversationId }, { headers: { Authorization: `Bearer ${token}` } });
      alert("Message forwarded successfully!");
      this.closeForwardMenu(messageId);
    },
    async forwardMessage(targetConversationId, messageId) {
      const message = this.messages.find(m => m.id === messageId);
      if (!message) return;
      const token = localStorage.getItem("token");
      const forwarderName = localStorage.getItem("name") || "Unknown";
      await axios.post(`/conversations/${this.conversationId}/message/${messageId}/forward`, { targetConversationId: targetConversationId, forwarderName: forwarderName }, { headers: { Authorization: `Bearer ${token}` } });
      alert("Message forwarded successfully!");
      this.closeForwardMenu(messageId);
    },
    getForwardButtonStyle(message) {
      return message.senderId === this.userToken ? { left: '-40px', top: 'calc(50% - 20px)' } : { right: '-40px', top: 'calc(50% - 20px)' };
    },
    getActionButtonStyle(message) {
      return message.senderId === this.userToken ? { left: '-40px', top: 'calc(50% + 5px)' } : { right: '-40px', top: 'calc(50% + 5px)' };
    }
  },
  mounted() {
    this.fetchMessages();
    this.pollIntervalId = setInterval(() => { this.fetchMessages(); }, 3000);
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
  display: flex;
  align-items: center;
  padding: 15px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
}
.chat-photo {
  width: 40px;
  height: 40px;
  margin-right: 10px;
}
.chat-photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
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
  border-radius: 10px;
  padding: 10px;
}
.message.self {
  margin-left: auto;
  background-color: #d1e7dd;
}
.message.other {
  background-color: #e0f2f1;
  padding-left: 45px;
  position: relative;
}
.sender-thumbnail {
  position: absolute;
  left: 10px;
  top: 10px;
  width: 30px;
  height: 30px;
}
.sender-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}
.message-content {
  position: relative;
  min-height: 40px;
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
  font-size: 0.8em;
}
.attachment-container {
  margin-top: 8px;
  width: 300px;
  height: 300px;
  overflow: hidden;
  border: 1px solid #ddd;
  border-radius: 8px;
}
.attachment-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
