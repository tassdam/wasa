<template>
    <div class="group-create-container">
      <h1 class="page-title">Create Group</h1>
      <div class="form-group">
        <label for="group-name">Group Name:</label>
        <input
          id="group-name"
          v-model="groupName"
          class="input-field"
          type="text"
          placeholder="Enter group name"
        />
      </div>
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
        <h2 class="results-title">Search Results:</h2>
        <template v-if="users.length > 0">
          <div v-for="user in users" :key="user.id" class="user-card">
            <h5 class="user-name">@{{ user.name }}</h5>
            <button
              class="add-button"
              @click="addUserToGroup(user)"
              :disabled="isUserAdded(user)"
            >
              {{ isUserAdded(user) ? 'Added' : 'Add' }}
            </button>
          </div>
        </template>
        <template v-else>
          <p class="no-results">No users found matching "{{ lastQuery }}"</p>
        </template>
      </div>
      <div v-if="selectedUsers.length > 0" class="selected-users-section">
        <h2 class="selected-title">Selected Members:</h2>
        <div class="selected-users">
          <span v-for="user in selectedUsers" :key="user.id" class="selected-user">
            @{{ user.name }}
            <button @click="removeUserFromGroup(user)" class="remove-button">Remove</button>
          </span>
        </div>
      </div>
      <div class="form-group">
        <label for="group-image">Group Image:</label>
        <input
          id="group-image"
          ref="fileInput"
          type="file"
          @change="handleImageUpload"
          accept="image/*"
        />
        <img v-if="previewImage" :src="previewImage" class="preview-image" alt="Group Image Preview"/>
      </div>
      <button class="create-button" @click="createGroup" :disabled="!canCreateGroup">
        Create Group
      </button>
    </div>
  </template>

  <script>
  import axios from "../services/axios";
  import LoadingSpinner from "../components/LoadingSpinner.vue";
  
  export default {
    name: "GroupCreateView",
    components: {
      LoadingSpinner,
    },
    data() {
      const token = localStorage.getItem("token");
      if (!token) {
        this.$router.push({ path: "/" });
        return;
      }
      return {
        groupName: "",
        query: "",
        lastQuery: "",
        users: [],
        loading: false,
        showResults: false,
        error: "",
        selectedUsers: [],
        previewImage: null,
        file: null,
      };
    },
    computed: {
      canCreateGroup() {
        return (
          this.groupName.trim() !== "" &&
          this.selectedUsers.length > 0 &&
          this.file !== null
        );
      },
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
          this.users = response.data.filter(user => user.id !== localStorage.getItem("token"));
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
      addUserToGroup(user) {
        if (!this.isUserAdded(user)) {
          this.selectedUsers.push(user);
        }
      },
      isUserAdded(user) {
        return this.selectedUsers.some((u) => u.id === user.id);
      },
      removeUserFromGroup(user) {
        this.selectedUsers = this.selectedUsers.filter((u) => u.id !== user.id);
      },
      handleImageUpload(event) {
        const file = event.target.files[0];
        if (file) {
          this.file = file;
          const reader = new FileReader();
          reader.onload = (e) => {
            this.previewImage = e.target.result;
          };
          reader.readAsDataURL(file);
        } else {
          this.previewImage = null;
          this.file = null;
        }
      },
      async createGroup() {
        if (!this.canCreateGroup) {
          alert("Please fill in all required fields and select a group image.");
          return;
        }
        this.loading = true;
        this.error = "";
        const formData = new FormData();
        formData.append("name", this.groupName);
        formData.append("image", this.file);
        formData.append("members", JSON.stringify([...this.selectedUsers.map(u => u.id), localStorage.getItem("token")]));
        try {
          await axios.post(`/groups`, formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            }
          });
          alert("Group created successfully!");
          this.$router.push(`/home`);
        } catch (err) {
          const status = err.response?.status;
          const reason = err.response?.data?.message || "Failed to create group.";
          this.error = `Status ${status}: ${reason}`;
        } finally {
          this.loading = false;
        }
      },
    },
  };
  </script>
  <style scoped>
  .group-create-container {
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
  
  .form-group {
    margin-bottom: 20px;
  }
  
  .input-field, .search-box {
    padding: 10px;
    border: 1px solid #ccc;
    border-radius: 5px;
    font-size: 16px;
    width: 100%;
    box-sizing: border-box;
  }
  
  .search-form {
    display: flex;
    justify-content: center;
    margin-bottom: 20px;
  }
  
  .search-button, .create-button {
    padding: 10px 20px;
    background-color: #007bff;
    color: #fff;
    border: none;
    border-radius: 5px;
    font-size: 16px;
    cursor: pointer;
    margin-left: 10px;
  }
  
  .search-button:hover, .create-button:hover {
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
    max-height: 300px; 
    overflow-y: auto;
    border: 1px solid #ccc;
    border-radius: 5px;
    padding: 10px;
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
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .user-name {
    font-size: 18px;
    color: #007bff;
  }
  
  .add-button, .remove-button {
    padding: 5px 10px;
    background-color: #28a745;
    color: #fff;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-size: 14px;
  }
  
  .add-button:hover, .remove-button:hover {
    background-color: #218838;
  }
  
  .add-button[disabled], .remove-button[disabled] {
    background-color: #ccc;
    cursor: not-allowed;
  }
  
  .selected-users-section {
    margin-top: 20px;
  }
  
  .selected-title {
    font-size: 20px;
    font-weight: bold;
    color: #444;
    margin-bottom: 10px;
  }
  
  .selected-users {
    display: flex;
    flex-wrap: wrap;
  }
  
  .selected-user {
    padding: 5px 10px;
    background-color: #e9ecef;
    border-radius: 5px;
    margin: 5px;
    font-size: 16px;
  }
  
  .preview-image {
    max-width: 200px;
    max-height: 200px;
    margin-top: 10px;
  }
  @media (max-width: 768px) {
  .results-section {
    max-height: 200px;
  }
}
  </style>