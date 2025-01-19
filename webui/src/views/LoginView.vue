<script>
export default {
  data() {
    localStorage.clear();
    return {
      errormsg: null,
      name: "", 
      profile: {
        id: "",
        name: "",
      },
    };
  },
  methods: {
    async loadDefaultPhoto() {
      try {
          const response = await fetch('/nopfp.jpg');
          const blob = await response.blob();
          const reader = new FileReader();
          return new Promise((resolve, reject) => {
              reader.onload = () => {
                  const base64String = reader.result.toString().split(',')[1];
                  resolve(base64String);
              };
              reader.onerror = reject;
              reader.readAsDataURL(blob);
          });
      } catch (error) {
          console.error('Error loading default photo:', error);
          return null;
      }
  },
    blobToBase64(blob) {
      return new Promise((resolve) => {
        const reader = new FileReader();
        reader.onloadend = () => resolve(reader.result);
        reader.readAsDataURL(blob);
      });
    },
    async doLogin() {
      if (this.name.trim() === "") {
        this.errormsg = "Name cannot be empty.";
        return;
      }
      try {
        const photoData = await this.loadDefaultPhoto();
        const response = await this.$axios.post("/session", {
          name: this.name,
          photo: photoData,
        }, {
          headers: {
            'Content-Type': 'application/json'
          }
        });
        if (response.data.identifier) {
          this.profile.id = response.data.identifier;
          this.profile.name = this.name; 
        } else {
          throw new Error("Unexpected server response. Missing 'identifier'.");
        }
        localStorage.setItem("token", this.profile.id);
        localStorage.setItem("name", this.profile.name);
        this.$router.push({ path: "/home" });
      } catch (e) {
        if (e.response && e.response.status === 400) {
          this.errormsg =
            "Form error, please check all fields and try again.";
        } else if (e.response && e.response.status === 500) {
          this.errormsg =
            "An internal error occurred. Please try again later.";
        } else {
          this.errormsg = e.toString();
        }
      }
    },
  },
};
</script>

<template>
  <div class="login-container">
    <h1 class="login-title">Welcome to WASAText</h1>
    <div class="input-group">
      <input
        type="text"
        id="name"
        v-model="name"
        class="login-input"
        placeholder="Insert your name to log in WASAText."
      />
      <button class="login-button" type="button" @click="doLogin">Login</button>
    </div>
    <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
  </div>
</template>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 100px auto;
  text-align: center;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 8px;
  background-color: #f9f9f9;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.login-title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #333;
}

.input-group {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.login-input {
  padding: 10px;
  width: 100%;
  margin-bottom: 10px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.login-button {
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 16px;
}

.login-button:hover {
  background-color: #0056b3;
}
</style>
