<template>
  <div class="search-container">
    <h1 class="page-title">Search People</h1>

    <!-- Search Form -->
    <form @submit.prevent="searchUsers" class="search-form">
      <input
        id="username"
        v-model="query"
        class="search-box"
        type="text"
        placeholder="Search by username"
      />
      <button class="search-button" type="submit">Search</button>
    </form>

    <!-- Error Message -->
    <div v-if="error" class="error-box">
      {{ error }}
    </div>

    <!-- Loading Spinner -->
    <div v-if="loading">
      <LoadingSpinner />
    </div>

    <!-- Search Results Section -->
    <div v-if="!loading && showResults" class="results-section">
      <h2 class="results-title">Results:</h2>
      <template v-if="users.length > 0">
        <div v-for="user in users" :key="user.id" class="user-card">
          <h5 @click="viewProfile(user.username)" class="user-name">
            @{{ user.name }}
          </h5>
          <button
            class="text-button"
            @click="navigateToConversation(user.id, user.name)"
          >
            Text
          </button>
        </div>
      </template>
      <template v-else>
        <p class="no-results">No users found matching "{{ lastQuery }}"</p>
      </template>
    </div>
  </div>
</template>

<script>
import axios from "../services/axios";
import LoadingSpinner from "../components/LoadingSpinner.vue";

export default {
  name: "SearchPeopleView",
  components: {
    LoadingSpinner,
  },
  data() {
    return {
      query: "",
      lastQuery: "",
      users: [],
      loading: false,
      showResults: false,
      error: "",
    };
  },
  methods: {
    async searchUsers() {
      if (!this.query.trim()) {
        this.error = "Please enter a valid search query.";
        this.showResults = false;
        return;
      }

      this.loading = true;
      this.error = "";
      this.users = [];
      this.showResults = false;

      try {
        const response = await axios.get(`/users/search`, {
          params: { username: this.query },
        });
        this.users = response.data;
        this.lastQuery = this.query;
        this.showResults = true;
      } catch (err) {
        const status = err.response?.status;
        const reason = err.response?.data?.message || "Failed to fetch users.";
        this.error = `Status ${status}: ${reason}`;
      } finally {
        this.loading = false;
      }
    },
    viewProfile(username) {
      this.$router.push(`/profile/${username}`);
    },
    navigateToConversation(recipientId, recipientName) {
      localStorage.setItem("recipientId", recipientId);
      localStorage.setItem("recipientName", recipientName);
      const senderId = localStorage.getItem("token"); // Get the sender's ID from local storage
      axios
        .post(`/conversations`, { senderId, recipientId })
        .then((response) => {
          const conversationId = response.data.conversationId; // Backend returns conversationId
          this.$router.push({
            path: `/conversations/${conversationId}`
          });
        })
        .catch((error) => {
          console.error("Error starting conversation:", error);
        });
    }
  },
};
</script>

<style scoped>
/* Container Styling */
.search-container {
  text-align: center;
  padding: 20px;
  max-width: 600px;
  margin: 0 auto;
}

/* Page Title */
.page-title {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #333;
}

/* Search Form */
.search-form {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.search-box {
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 5px;
  font-size: 16px;
  width: 300px;
}

.search-button {
  padding: 10px 20px;
  background-color: #007bff;
  color: #fff;
  border: none;
  border-radius: 5px;
  font-size: 16px;
  margin-left: 10px;
  cursor: pointer;
}

.search-button:hover {
  background-color: #0056b3;
}

/* Error Message */
.error-box {
  background-color: #f8d7da;
  color: #842029;
  border: 1px solid #f5c2c7;
  border-radius: 5px;
  padding: 10px;
  margin: 20px 0;
  text-align: center;
}

/* Results Section */
.results-section {
  margin-top: 20px;
}

.results-title {
  font-size: 24px;
  font-weight: bold;
  color: #444;
  margin-bottom: 10px;
}

.user-card {
  padding: 10px;
  margin: 10px 0;
  background-color: #f9f9f9;
  border: 1px solid #ccc;
  border-radius: 5px;
  font-size: 18px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-card:hover {
  background-color: #e9ecef;
}

.user-name {
  margin: 0;
  font-size: 18px;
  color: #007bff;
}

.user-name:hover {
  text-decoration: underline;
}

/* Text Button */
.text-button {
  padding: 5px 10px;
  background-color: #28a745;
  color: #fff;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
}

.text-button:hover {
  background-color: #218838;
}

/* No Results Message */
.no-results {
  font-size: 16px;
  color: #666;
  margin-top: 20px;
}
</style>
