<template>
  <div>
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">{{ userName }}, search and text people!</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">Refresh</button>
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="logOut">Log Out</button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newGroup">New group</button>
        </div>
      </div>
    </div>

    <div class="search-container">
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
      <div v-if="error" class="error-box">
        {{ error }}
      </div>
      <div v-if="loading">
        <LoadingSpinner />
      </div>
      <div v-if="!loading && showResults" class="results-section">
        <h2 class="results-title">Results:</h2>
        <template v-if="users.length > 0">
          <div v-for="user in users" :key="user.id" class="user-card">
            <h5 class="user-name">
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
      userName: localStorage.getItem("name"),
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
        const response = await axios.get(`/search`, {
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
    navigateToConversation(recipientId, recipientName) {
      localStorage.setItem("conversationName", recipientName);
      const senderId = localStorage.getItem("token");
      axios
        .post(`/conversations`, { senderId, recipientId })
        .then((response) => {
          const conversationId = response.data.conversationId;
          this.$router.push({
            path: `/conversations/${conversationId}`
          });
        })
        .catch((error) => {
          console.error("Error starting conversation:", error);
        });
    },
    refresh() {
      this.$router.push({ path: "/search" });
    },
    logOut() {
      localStorage.clear();
      this.$router.push({ path: "/" });
    },
    newGroup() {
      this.$router.push({ path: "/new-group" });
    }
  },
  mounted() {
    const token = localStorage.getItem("token");
    if (!token) {
      this.$router.push({ path: "/" });
    }
  }
};
</script>

<style scoped>
.search-container {
  text-align: center;
  padding: 20px;
  max-width: 600px;
  margin: 0 auto;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #333;
}

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

.error-box {
  background-color: #f8d7da;
  color: #842029;
  border: 1px solid #f5c2c7;
  border-radius: 5px;
  padding: 10px;
  margin: 20px 0;
  text-align: center;
}

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

.no-results {
  font-size: 16px;
  color: #666;
  margin-top: 20px;
}
</style>