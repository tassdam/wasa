<script>
export default {
  data() {
    return {
      errormsg: null,
      loading: false,
      some_data: null,
      conversations: []  // We'll store the list of convos here
    }
  },
  methods: {
    async loadConversations() {
      this.errormsg = null
      this.loading = true
      try {
        // If user is not logged in, you might not have 'token' in localStorage
        const token = localStorage.getItem("token")
        if (!token) {
          // Show a message but do NOT redirect or vanish the UI
          this.errormsg = "No token found - are you sure you are logged in?"
          this.loading = false
          return
        }

        // Attempt the request
        let response = await this.$axios.get("/conversations", {
          headers: {
            Authorization: "Bearer " + token
          }
        })
        this.conversations = response.data || []
      } catch (e) {
        console.error("loadConversations error:", e)
        // Show the error in a visible spot
        this.errormsg = "Failed to load conversations: " + e.toString()
      }
      this.loading = false
    },

    // Optional placeholders
    exportList() {
      console.log("Export triggered")
    },
    newItem() {
      console.log("New item triggered")
    }
  },
  mounted() {
    // 1) run your existing refresh
    this.refresh()

    // 2) load convos automatically
    this.loadConversations()
  }
}
</script>

<template>
  <div>
    <!-- Top bar -->
    <div
      class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
    >
      <h1 class="h2">Home page</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button
            type="button"
            class="btn btn-sm btn-outline-secondary"
            @click="refresh"
          >
            Refresh
          </button>
          <button
            type="button"
            class="btn btn-sm btn-outline-secondary"
            @click="exportList"
          >
            Export
          </button>
        </div>
        <div class="btn-group me-2">
          <button
            type="button"
            class="btn btn-sm btn-outline-primary"
            @click="newItem"
          >
            New
          </button>
        </div>
      </div>
    </div>

    <!-- Show any error message -->
    <ErrorMsg v-if="errormsg" :msg="errormsg" />

    <!-- Show the "some_data" from refresh() if it exists -->
    <div v-if="some_data">
      <p>some_data: {{ some_data }}</p>
    </div>

    <!-- Show conversation list (if any) -->
    <div>
      <h3>My Conversations</h3>
      <!-- If loading is true, you can optionally show some spinner -->
      <p v-if="loading">Loading...</p>

      <!-- If not loading but no convos, show a note -->
      <div v-else-if="conversations.length === 0">
        <p>No conversations found.</p>
      </div>

      <!-- Otherwise list them -->
      <ul v-else>
        <li
          v-for="conv in conversations"
          :key="conv.id"
        >
          <strong>{{ conv.name }}</strong> - members: {{ conv.members.join(", ") }}
          <div v-if="conv.lastMessage">
            Last message: {{ conv.lastMessage.content }} 
            at {{ conv.lastMessage.timestamp }}
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<style>
/* optional styling */
</style>
