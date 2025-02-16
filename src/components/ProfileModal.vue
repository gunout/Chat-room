<script>
import VideoFlag from '../lib/VideoFlag';

export default {
    props: {
        visible: Boolean,
        jwt: String, // caller's JWT token for authorization
        user: Object, // the user we are viewing
        username: String, // the local user
        isViewerOp: Boolean, // the viewer is an operator (show buttons)
        websiteUrl: String,
        isDnd: Boolean,
        isMuted: Boolean,
        isBooted: Boolean,
        profileWebhookEnabled: Boolean,
        vipConfig: Object,  // VIP config settings for BareRTC
    },
    data() {
        return {
            busy: false,

            // Profile data
            profileFields: [],

            // Ban account data
            banModalVisible: false,
            banReason: "",
            banDuration: 24,

            // Error messaging from backend
            error: null,
        };
    },
    watch: {
        visible() {
            if (this.visible) {
                this.refresh();
            } else {
                this.profileFields = [];
                this.error = null;
                this.busy = false;
            }
        }
    },
    computed: {
        profileURL() {
            if (this.user.profileURL) {
                return this.urlFor(this.user.profileURL);
            }
            return null;
        },
        avatarURL() {
            if (this.user.avatar) {
                return this.urlFor(this.user.avatar);
            }
            return null;
        },
        nickname() {
            if (this.user.nickname) {
                return this.user.nickname;
            }
            return this.user.username;
        },
        isOnBlueCam() {
            // User is broadcasting a cam and is not NSFW.
            if ((this.user.video & VideoFlag.Active) && !(this.user.video & VideoFlag.NSFW)) {
                return true;
            }
            return false;
        },
        isOnCamera() {
            // User's camera is enabled.
            return (this.user.video & VideoFlag.Active);
        },
    },
    methods: {
        refresh() {
            if (!this.profileWebhookEnabled) return;
            if (!this.user || !this.user?.username) return;
            this.busy = true;
            return fetch("/api/profile", {
                method: "POST",
                mode: "same-origin",
                cache: "no-cache",
                credentials: "same-origin",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    "JWTToken": this.jwt,
                    "Username": this.user?.username,
                }),
            })
            .then((response) => response.json())
            .then((data) => {
                if (data.Error) {
                    this.error = data.Error;
                    return;
                }

                if (data.ProfileFields != undefined) {
                    this.profileFields = data.ProfileFields;
                }
            }).catch(resp => {
                this.error = resp;
            }).finally(() => {
                this.busy = false;
            })
        },

        cancel() {
            this.$emit("cancel");
        },

        openProfile() {
            let url = this.profileURL;
            if (url) {
                window.open(url);
            }
        },

        openDMs() {
            this.cancel();
            this.$emit('send-dm', {
                username: this.user.username,
            });
        },

        muteUser() {
            this.$emit('mute-user', this.user.username);
        },

        bootUser() {
            this.$emit('boot-user', this.user.username);
        },

        // Operator commands (may be rejected by server if not really Op)
        markNsfw() {
            if (!window.confirm("Mark this user's webcam as 'Explicit'?")) return;
            this.$emit('send-command', `/nsfw ${this.user.username}`);

            // Close the modal immediately: our view of the user's cam data is a copy
            // and we can't follow the current value.
            this.cancel();
        },
        cutCamera() {
            if (!window.confirm("Make this user stop broadcasting their camera?")) return;
            this.$emit('send-command', `/cut ${this.user.username}`);
            this.cancel();
        },
        kickUser() {
            if (!window.confirm("Really kick this user from the chat room?")) return;
            this.$emit('send-command', `/kick ${this.user.username}`);
        },
        banUser() {
            this.banModalVisible = true;
            this.banReason = "";
            this.banDuration = 24;
            window.requestAnimationFrame(() => {
                let reason = document.querySelector("#ban_reason");
                if (reason) {
                    reason.focus();
                }
            });
        },
        confirmBan() {
            // Send the ban command.
            this.$emit('send-command', `/ban ${this.user.username} ${this.banDuration}`);

            // Also send an admin report to the main website.
            this.$emit('report', {
                message: {
                    channel: `n/a`,
                    username: this.user.username,
                    at: new Date(),
                    message: 'Ban reason: ' + this.banReason,
                },
                classification: 'User banned by admin',
                comment: `A chat admin has banned ${this.user.username} from the chat room!\n\n` +
                    `* Chat admin: <a href="/u/${this.username}">${this.username}</a>\n` +
                    `* Reason: ${this.banReason}\n` +
                    `* Duration: ${this.banDuration} hours`,
            });

            this.banModalVisible = false;
            this.cancel();
        },

        urlFor(url) {
            // Prepend the base websiteUrl if the given URL is relative.
            if (url.match(/^https?:/i)) {
                return url;
            }
            return this.websiteUrl.replace(/\/+$/, "") + url;
        },
    },
}
</script>

<template>
    <!-- Profile Card Modal -->
    <div class="modal" :class="{ 'is-active': visible }">
        <div class="modal-background" @click="cancel()"></div>
        <div class="modal-content">
            <div class="card">
                <header class="card-header has-background-success">
                    <p class="card-header-title">Profile Card</p>
                </header>
                <div class="card-content">

                    <!-- Avatar and name/username media -->
                    <div class="media mb-0">
                        <div class="media-left">
                            <a :href="profileURL"
                                @click.prevent="openProfile()"
                                :class="{ 'cursor-default': !profileURL }">
                                <figure class="image is-96x96">
                                    <img v-if="avatarURL"
                                        :src="avatarURL">
                                    <img v-else src="/static/img/shy.png">
                                </figure>
                            </a>
                        </div>
                        <div class="media-content">
                            <strong>
                                <!-- User nickname/display name -->
                                {{ nickname }}
                            </strong>
                            <div>
                                <small>
                                    <a v-if="profileURL"
                                        :href="profileURL" target="_blank"
                                        class="has-text-grey">
                                        @{{ user.username }}
                                    </a>
                                    <span v-else class="has-text-grey">@{{ user.username }}</span>
                                </small>
                            </div>

                            <!-- User badges -->
                            <div v-if="user.op || user.vip || user.emoji" class="mt-4">
                                <!-- Emoji icon -->
                                <span v-if="user.emoji" class="mr-2">
                                    {{ user.emoji }}
                                </span>

                                <!-- Operator? -->
                                <span v-if="user.op" class="tag is-warning is-light mr-2">
                                    <i class="fa fa-peace mr-1"></i> Operator
                                </span>

                                <!-- VIP? -->
                                <span v-if="vipConfig && user.vip" class="tag is-success is-light mr-2"
                                    :title="vipConfig.Name">
                                    <i class="mr-1" :class="vipConfig.Icon"></i>
                                    {{ vipConfig.Name }}
                                </span>
                            </div>
                        </div>
                    </div>

                    <!-- Action buttons -->
                    <div v-if="user.username !== username" class="mt-4">
                        <!-- DMs button -->
                        <button type="button"
                            class="button is-small px-2 mb-1"
                            @click="openDMs()"
                            :title="isDnd ? 'This person is not accepting new DMs' : 'Open a Direct Message (DM) thread'"
                            :disabled="isDnd">
                            <i class="fa mr-1" :class="{'fa-comment': !isDnd, 'fa-comment-slash': isDnd}"></i>
                            Direct Message
                        </button>

                        <!-- Mute button -->
                        <button type="button"
                            class="button is-small px-2 ml-1 mb-1"
                            @click="muteUser()" title="Mute user">
                            <i class="fa fa-comment-slash mr-1" :class="{
                                'has-text-success': isMuted,
                                'has-text-danger': !isMuted
                            }"></i>
                            {{ isMuted ? "Unmute" : "Mute" }} Messages
                        </button>

                        <!-- Boot button -->
                        <button type="button"
                            class="button is-small px-2 ml-1 mb-1"
                            @click="bootUser()" title="Boot user off your webcam">
                            <i class="fa fa-user-xmark mr-1" :class="{
                                'has-text-danger': !isBooted,
                                'has-text-success': isBooted,
                            }"></i>
                            {{  isBooted ? 'Allow to watch my webcam' : "Don't allow to watch my webcam" }}
                        </button>

                        <!-- Admin actions -->
                        <div v-if="isViewerOp" class="mt-1">
                            <!-- Mark camera NSFW -->
                            <button v-if="isOnBlueCam"
                                type="button"
                                class="button is-small is-outlined is-danger has-text-dark px-2 mr-1 mb-1"
                                @click="markNsfw()" title="Mark their camera as Explicit (red).">
                                <i class="fa fa-video mr-1 has-text-danger"></i>
                                Mark camera as Explicit
                            </button>

                            <!-- Cut camera -->
                            <button v-if="isOnCamera"
                                type="button"
                                class="button is-small is-outlined is-danger has-text-dark px-2 mr-1 mb-1"
                                @click="cutCamera()" title="Turn their camera off.">
                                <i class="fa fa-stop mr-1 has-text-danger"></i>
                                Cut camera
                            </button>

                            <!-- Kick user -->
                            <button type="button"
                                class="button is-small is-outlined is-danger has-text-dark px-2 mr-1 mb-1"
                                @click="kickUser()" title="Kick this user from the chat room.">
                                <i class="fa fa-shoe-prints mr-1 has-text-danger"></i>
                                Kick from the room
                            </button>

                            <!-- Ban user -->
                            <button type="button"
                                class="button is-small is-outlined is-danger has-text-dark px-2 mb-1"
                                @click="banUser()" title="Ban this user from the chat room for 24 hours.">
                                <i class="fa fa-clock mr-1 has-text-danger"></i>
                                Ban from chat
                            </button>
                        </div>
                    </div>

                    <!-- Profile Fields spinner/error -->
                    <div class="notification is-info is-light p-2 my-2" v-if="busy">
                        <i class="fa fa-spinner fa-spin mr-2"></i>
                        Loading profile details...
                    </div>
                    <div class="notification is-danger is-light p-2 my-2" v-else-if="error">
                        <i class="fa fa-exclamation-triangle mr-2"></i>
                        Error loading profile details:
                        {{ error }}
                    </div>

                    <!-- Profile Fields -->
                    <div class="columns is-multiline is-mobile mt-3"
                        v-else-if="profileFields.length > 0">
                        <div class="column py-1"
                            v-for="(field, i) in profileFields"
                            v-bind:key="field.Name"
                            :class="{'is-half': i < profileFields.length-1}">
                            <strong>{{ field.Name }}:</strong>
                            {{ field.Value }}
                        </div>
                    </div>
                </div>
                <footer class="card-footer">
                    <a :href="profileURL" target="_blank"
                        v-if="profileURL" class="card-footer-item"
                        @click="cancel()">
                        Full profile <i class="fa fa-external-link ml-2"></i>
                    </a>
                    <a href="#" @click.prevent="cancel()" class="card-footer-item">
                        Close
                    </a>
                </footer>
            </div>
        </div>
    </div>

    <!-- Ban User Modal (for chat admins) -->
    <div class="modal" :class="{ 'is-active': banModalVisible }">
        <div class="modal-background" @click="banModalVisible = false"></div>
        <div class="modal-content">
            <form @submit.prevent="confirmBan">
                <div class="card">
                    <header class="card-header has-background-danger">
                        <p class="card-header-title">Ban User</p>
                    </header>
                    <div class="card-content">
                        <div class="field">
                            <label class="label" for="ban_reason">Reason for the ban:</label>
                            <input type="text" class="input"
                                id="ban_reason"
                                placeholder="Please describe why this user will be banned."
                                v-model="banReason"
                                required>
                            <p class="help">
                                This reason is NOT shown to the banned user, but will be sent to the main website
                                in an admin report so that it may be documented in this user's history.
                            </p>
                        </div>

                        <div class="field">
                            <label class="label" for="ban_duration">How long for the ban? (1-96 hours)</label>
                            <input type="number" min="1" max="96" v-model="banDuration" class="input">
                        </div>

                        <div class="field has-text-centered">
                            <button type="submit" class="button is-danger">
                                Confirm Ban
                            </button>
                            <a href="#" @click.prevent="banModalVisible = false" class="button ml-2">
                                Cancel
                            </a>
                        </div>
                    </div>
                </div>
            </form>
        </div>
    </div>
</template>

<style scoped>
</style>
