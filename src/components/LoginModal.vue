<script>
export default {
    props: {
        visible: Boolean,
    },
    data() {
        return {
            username: '',
        };
    },
    methods: {
        signIn() {
            this.$emit('signIn', this.username);
        }
    }
}
</script>

<template>
    <div class="modal" :class="{ 'is-active': visible }">
        <div class="modal-background"></div>
        <div class="modal-content">
            <div class="card">
                <header class="card-header has-background-info">
                    <p class="card-header-title">Sign In</p>
                </header>
                <div class="card-content">
                    <form @submit.prevent="signIn()">

                        <div v-if="autoLogin" class="content">
                            <p>
                                Welcome to <span v-html="config.branding"></span>! Please just click on the "Enter Chat"
                                button below to log on. Your username has been pre-filled from the website that
                                sent you here.
                            </p>

                            <p>
                                This dialog box is added as an experiment to see whether it
                                helps iOS devices (iPads and iPhones) to log in to the chat more reliably, by
                                having you interact with the page before it connects to the server. Let us
                                know in chat if your iPhone or iPad is able to log in this way!
                            </p>
                        </div>

                        <div class="field">
                            <label class="label">Username</label>
                            <input class="input" v-model="username" placeholder="Username" autocomplete="off" autofocus
                                :disabled="autoLogin" required>
                        </div>

                        <div class="field">
                            <div class="control">
                                <button class="button is-link">Enter Chat</button>
                            </div>
                        </div>

                    </form>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
</style>
