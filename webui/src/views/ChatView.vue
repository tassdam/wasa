<template>
    <div class="chat-container">
      <h1>Conversation with {{ userId }}</h1>
      <p>This is the chat page for the user with UUID: {{ userId }}.</p>
  
      <!-- Placeholder for chat interface -->
      <div class="chat-box">
        <input
          v-model="message"
          class="message-input"
          type="text"
          placeholder="Type a message..."
        />
        <button class="send-button" @click="sendMessage">Send</button>
      </div>
    </div>
  </template>
  
  <script>
  export default {
    name: "ChatView",
    props: ["uuid"],
    data() {
      return {
        userId: this.$route.params.uuid, // Get the user's UUID from the route
        message: "",
      };
    },
    methods: {
      sendMessage() {
        if (!this.message.trim()) {
          alert("Please type a message before sending.");
          return;
        }
        console.log(`Message to ${this.userId}: ${this.message}`);
        this.message = ""; // Clear the input after sending
      },
    },
  };
  </script>
  
  <style scoped>
  .chat-container {
    padding: 20px;
    text-align: center;
  }
  
  .chat-box {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
  
  .message-input {
    padding: 10px;
    width: 300px;
    border: 1px solid #ccc;
    border-radius: 5px;
  }
  
  .send-button {
    padding: 10px 20px;
    margin-left: 10px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
  }
  
  .send-button:hover {
    background-color: #0056b3;
  }
  </style>
  