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
            <strong>{{ message.senderId === userToken ? 'You' : (message.senderName || 'Unknown Sender') }}</strong>:
            {{ message.content }}
          </p>
          <small>{{ formatTimestamp(message.timestamp) }}</small>
          <!-- Reactions Display -->
          <div v-if="messageOptions[message.id]?.reactions && messageOptions[message.id].reactions.length">
            <span v-for="(reaction, index) in messageOptions[message.id].reactions" :key="index">
              {{ reaction.emoji }}
            </span>
          </div>
          <!-- Forward Button -->
          <button class="forward-button" @click.stop="showForwardOptions(message.id)">→</button>
          <!-- Delete Button for self messages -->
          <button v-if="message.senderId === userToken" class="delete-button" @click.stop="deleteMessage(message)">✖</button>
          <!-- Comment Emojis -->
          <div v-if="messageOptions[message.id]?.showCommentEmojis" class="comment-emojis" @click.stop>
            <button class="emoji-button" @click="sendReaction('😄', message.id)">😄</button>
            <button class="emoji-button" @click="sendReaction('😅', message.id)">😅</button>
            <button class="emoji-button" @click="sendReaction('😍', message.id)">😍</button>
            <button class="emoji-button" @click="sendReaction('🤔', message.id)">🤔</button>
            <button class="emoji-button" @click="sendReaction('😢', message.id)">😢</button>
          </div>
          <!-- Forward Options -->
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
      showCommentEmojis: false,
      emojis: ['😄', '😅', '😍', '🤔', '😢'],
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
        this.messages = response.data.messages || [];
        this.$nextTick(() => {
          this.scrollToBottom();
        });
      } catch (error) {
        console.error("Failed to fetch messages:", error);
        alert("Failed to load messages. Please try again later.");
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
          showCommentEmojis: false,
          forwardConversations: [],
          selectedConversationId: null,
          reactions: [],
        };
        this.fetchForwardConversations(messageId);
      } else {
        this.messageOptions[messageId].showForwardMenu = !this.messageOptions[messageId].showForwardMenu;
      }
    },
    showCommentEmojis(messageId) {
      this.closeAllMenus();
      if (!this.messageOptions[messageId]) {
        this.messageOptions[messageId] = {
          showForwardMenu: false,
          showCommentEmojis: true,
          reactions: [],
        };
      } else {
        this.messageOptions[messageId].showCommentEmojis = !this.messageOptions[messageId].showCommentEmojis;
      }
    },
    closeForwardMenu(messageId) {
      if (this.messageOptions[messageId]) {
        this.messageOptions[messageId].showForwardMenu = false;
      }
    },
    closeCommentEmojis(messageId) {
      if (this.messageOptions[messageId]) {
        this.messageOptions[messageId].showCommentEmojis = false;
      }
    },
    closeAllMenus() {
      for (const id in this.messageOptions) {
        this.messageOptions[id].showForwardMenu = false;
        this.messageOptions[id].showCommentEmojis = false;
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
        const response = await axios.get('users/me/conversations', {
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
      if (!targetConversationId) {
        alert("Please select a conversation to forward the message.");
        return;
      }
      const message = this.messages.find(m => m.id === messageId);
      if (!message) return;
      try {
        const token = localStorage.getItem("token");
        await axios.post(
          `/conversations/${targetConversationId}/messages/forward`,
          { sourceMessageId: message.id },
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
    sendReaction(emoji, messageId) {
      const message = this.messages.find(m => m.id === messageId);
      if (!message) return;
      try {
        const token = localStorage.getItem("token");
        axios.post(
          `/conversations/${this.conversationId}/messages/${messageId}/comments`,
          { emoji: emoji },
          {
            headers: { Authorization: `Bearer ${token}` },
          }
        ).then(() => {
          this.messageOptions[messageId].reactions.push({ emoji: emoji });
          this.closeCommentEmojis(messageId);
        }).catch(error => {
          console.error("Failed to send reaction:", error);
          alert("Failed to send reaction. Please try again.");
        });
      } catch (error) {
        console.error("Failed to send reaction:", error);
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

.forward-button, .delete-button {
  background-color: #00000000;
  border: none;
  position: absolute;
  top: 1px;
  right: 1px;
  cursor: pointer;
}

.delete-button {
  right: 20px;
}

.message-content:hover .forward-button, .message-content:hover .delete-button {
  display: block;
}

.forward-options, .comment-emojis {
  position: absolute;
  top: 30px;
  right: 0;
  background-color: #ffffff62;
  border-radius: 5px;
  padding: 5px;
}

.comment-emojis {
  display: flex;
  flex-direction: row;
}

.emoji-button {
  background-color: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  margin-right: 5px;
}

.emoji-button:hover {
  background-color: #cccccc;
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
  word-wrap: break-word;
  word-break: break-all;
  white-space: normal;
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

.forward-select {
  width: 100%;
  padding: 5px;
  border: 1px solid #ccc;
  border-radius: 3px;
  background-color: #ffffffcd;
  cursor: pointer;
  margin-bottom: 5px;
}

.forward-select:focus {
  outline: none;
  border-color: #999;
}

.button-style {
  background-color: #ffffff62;
  border: none;
  padding: 5px 10px;
  border-radius: 3px;
  cursor: pointer;
  margin-top: 5px;
}

.button-style:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}
</style>




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