<template>
    <div class="groupInfo-container">
      <div class="groupInfo-header">
        <div class="groupPhoto-container">
          <img v-if="groupPhoto" :src="groupPhoto" alt="Group Photo" class="group-photo" />
        </div>
        <div class="groupName-container">
          <h1 class="groupName">{{ groupName }}</h1>
          <div class="update-groupName-section">
            <input
              v-model="newGroupName"
              placeholder="Enter new group name"
              maxlength="16"
              minlength="3"
            />
            <button
              @click="updateGroupName"
              :disabled="!newGroupName || newGroupName === groupName"
            >
              Update Group Name
            </button>
          </div>
          <div class="update-groupPhoto-section">
            <input type="file" @change="handleGroupPhotoUpload" accept="image/*" />
            <button @click="updateGroupPhoto" :disabled="!newGroupPhoto">Update Group Photo</button>
          </div>
          <div class="leave-group-section">
            <button class="leave-button" @click="leaveGroup">
              Leave Group
            </button>
          </div>
        </div>
      </div>
      <ErrorMsg v-if="errormsg" :msg="errormsg" />
    </div>
  </template>
  
  <script>
  import axios from "../services/axios";
  import ErrorMsg from "../components/ErrorMsg.vue";
  
  export default {
    name: "GroupEditView",
    components: {
      ErrorMsg,
    },
    data() {
      return {
        groupId: this.$route.params.uuid,
        groupName: localStorage.getItem("groupName"),
        groupPhoto: null,
        newGroupName: "", 
        newGroupPhoto: null, 
        errormsg: null, 
      };
    },
    methods: {
        async fetchGroupProfile() {
            try {
                const token = localStorage.getItem("token");
                if (!token) {
                    this.$router.push({ path: "/" });
                    return;
                }
                const response = await axios.get(`/groups/${this.groupId}`, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });
                const groupPhoto  = response.data.groupPhoto;
                this.groupPhoto = groupPhoto ? `data:image/*;base64,${groupPhoto}` : null;
            } catch (error) {
                console.error("Failed to fetch user profile:", error);
                this.errormsg = "Failed to load user profile. Please try again later.";
            }
        },
      handleGroupPhotoUpload(event) {
        const file = event.target.files[0];
        if (file) {
          this.newGroupPhoto = file;
        }
      },
      async updateGroupPhoto() {
        if (!this.newGroupPhoto) return;
        try {
          const token = localStorage.getItem("token");
          const formData = new FormData();
          formData.append("photo", this.newGroupPhoto);
          await axios.put(`/groups/${this.groupId}/setGroupPhoto`, formData, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
          alert("Group photo updated successfully!");
          this.newGroupPhoto = null;
          this.fetchGroupProfile(); 
        } catch (error) {
          console.error("Failed to update group photo:", error);
          this.errormsg = "Failed to update group photo. Please try again.";
        }
      },
      async updateGroupName() {
        if (!this.newGroupName || this.newGroupName === this.GroupName) return;
        try {
          const token = localStorage.getItem("token");
            await axios.put(
            `/groups/${this.groupId}/setGroupName`,
            { groupName: this.newGroupName },
            {
              headers: {
              Authorization: `Bearer ${token}`,
              },
            }
            );
          alert("Group name updated successfully!");
          localStorage.setItem("groupName", this.newGroupName) 
          this.groupName = this.newGroupName;
          this.newGroupName = "";
        } catch (error) {
          console.error("Failed to update group name:", error);
          this.errormsg = "Failed to update group name. Please try again.";
        }
      },
      async leaveGroup() {
        if (!confirm('Are you sure you want to leave this group?')) {
          return;
        }
        
        try {
          const token = localStorage.getItem("token");
          await axios.delete(`/groups/${this.groupId}/leaveGroup`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
          this.$router.push({ path: "/groups" });
        } catch (error) {
          console.error("Failed to leave group:", error);
          this.errormsg = "Failed to leave group. Please try again.";
        }
      },  
    },
    mounted() {
      this.fetchGroupProfile();
    },
  };
  </script>
  
  <style scoped>
  .groupInfo-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;
  }
  
  .groupInfo-header {
    display: flex;
    align-items: flex-start;
    gap: 20px;
    width: 100%;
    max-width: 800px;
  }
  
  .groupPhoto-container {
    flex: 0 0 auto;
    width: 120px;
    height: 120px;
    border-radius: 50%;
    overflow: hidden;
    border: 1px solid #ccc;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #f5f5f5;
  }
  
  .group-photo {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .no-photo-placeholder {
    color: #aaa;
    font-size: 14px;
  }
  
  .groupName-container {
    flex: 1;
  }
  
  .groupName {
    margin: 0;
    font-size: 24px;
    font-weight: bold;
  }
  
  .update-groupName-section,
  .update-photo-section {
    margin-top: 10px;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  
  input {
    padding: 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
  }
  
  button {
    padding: 8px 12px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  
  button:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
  
  button:hover:not(:disabled) {
    background-color: #0056b3;
  }
  </style>
  