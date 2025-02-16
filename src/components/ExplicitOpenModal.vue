<script>
export default {
    props: {
        visible: Boolean,
        user: Object,
    },
    data() {
        return {
            dontShowAgain: false,
        };
    },
    methods: {
        accept() {
            if (this.dontShowAgain) {
                this.$emit('dont-show-again');
            }
            this.$emit('accept');
        },
        cancel() {
            this.$emit('cancel');
        },
    }
}
</script>

<template>
    <!-- NSFW Modal: before user views a NSFW camera the first time -->
    <div class="modal" :class="{ 'is-active': visible }">
        <div class="modal-background"></div>
        <div class="modal-content">
            <div class="card">
                <header class="card-header has-background-info">
                    <p class="card-header-title">This camera may contain Explicit content</p>
                </header>
                <div class="card-content">
                    <p class="block">
                        This camera has been marked as "Explicit/<abbr title="Not Safe For Work">NSFW</abbr>" and may
                        contain displays of sexuality. If you do not want to see this, look for cameras with
                        a <span class="button is-small is-info is-outlined px-1"><i class="fa fa-video"></i></span>
                        blue icon rather than the <span class="button is-small is-danger is-outlined px-1"><i
                                class="fa fa-video"></i></span>
                        red ones.
                    </p>

                    <div class="field">
                        <label class="checkbox">
                            <input type="checkbox" v-model="dontShowAgain">
                            Don't show this message again
                        </label>
                    </div>

                    <div class="field">
                        <div class="control has-text-centered">
                            <button type="button" class="button is-link mr-4"
                                @click="accept()">
                                Open webcam
                            </button>
                            <button type="button" class="button"
                                @click="cancel()">
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
</style>
