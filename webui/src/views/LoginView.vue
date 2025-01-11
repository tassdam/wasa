<script>
export default {
    components: {},
    data: function () {
        return {
            errormsg: null,
            name: "",
            profile: {
                id: 0,
                name: "",
            },
        }
    },
    methods: {
        async doLogin() {
            if (this.name == "") {
                this.errormsg = "Username cannot be empty.";
            } else {
                try {
                    let response = await this.$axios.post("/session", { name: this.name })
                    this.profile = response.data
                    localStorage.setItem("token", this.profile.id);
                    localStorage.setItem("name", this.profile.name);
                    this.$router.push({ path: '/home' })
                } catch (e) {
                    if (e.response && e.response.status === 400) {
                        this.errormsg = "Form error, please check all fields and try again. If you think that this is an error, write an e-mail to us.";
                        this.detailedmsg = null;
                    } else if (e.response && e.response.status === 500) {
                        this.errormsg = "An internal error occurred. We will be notified. Please try again later.";
                        this.detailedmsg = e.toString();
                    } else {
                        this.errormsg = e.toString();
                        this.detailedmsg = null;
                    }
                }
            }

        }
    },
    mounted() {

    }

}
</script>

<template>
    <div
        class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
        <h1 class="h2">Welcome to WASAText</h1>
    </div>
    <div class="input-group mb-3">
        <input type="text" id="name" v-model="name" class="form-control"
            placeholder="Insert a username to log in WASAText." aria-label="Your username"
            aria-describedby="basic-addon2">
        <div class="input-group-append">
            <button class="btn btn-success" type="button" @click="doLogin">Login</button>
        </div>
    </div>
    <ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
</template>

<style>

</style>