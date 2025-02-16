<script>
import VideoFlag from '../lib/VideoFlag';

export default {
    props: {
        user: Object,       // User object of the Message author
        username: String,   // current username logged in
        websiteUrl: String, // Base URL to website (for profile/avatar URLs)
        isDnd: Boolean,     // user is not accepting DMs
        isMuted: Boolean,   // user is muted by current user
        isBlocked: Boolean, // user is blocked on your main website (can't be unmuted)
        isBooted: Boolean,  // user is booted by current user
        vipConfig: Object,  // VIP config settings for BareRTC
        isOp: Boolean,      // current user is operator (can always DM)
        isVideoNotAllowed: Boolean,  // whether opening this camera is not allowed
        videoIconClass: String,      // CSS class for the open video icon
        isWatchingTab: Boolean, // is the "Watching" tab (replace video button w/ boot)
        statusMessage: Object, // StatusMessage controller
    },
    data() {
        return {
            VideoFlag: VideoFlag,
        };
    },
    computed: {
        profileURL() {
            if (this.user.profileURL) {
                return this.urlFor(this.user.profileURL);
            }
            return null;
        },
        profileButtonClass() {
            let result = "";

            // VIP background.
            if (this.user.vip) {
                result = "has-background-vip ";
            }

            let gender = (this.user.gender || "").toLowerCase();
            if (gender.indexOf("m") === 0) {
                return result + "has-text-gender-male";
            } else if (gender.indexOf("f") === 0) {
                return result + "has-text-gender-female";
            } else if (gender.length > 0) {
                return result + "has-text-gender-other";
            }
            return "";
        },
        videoButtonClass() {
            let result = "";

            // VIP background if their cam is set to VIPs only
            if ((this.user.video & VideoFlag.Active) && (this.user.video & VideoFlag.VipOnly)) {
                result = "has-background-vip ";
            }

            // Colors and/or cursors.
            if ((this.user.video & VideoFlag.Active) && (this.user.video & VideoFlag.NSFW)) {
                result += "is-danger is-outlined";
            } else if ((this.user.video & VideoFlag.Active) && !(this.user.video & VideoFlag.NSFW)) {
                result += "is-link is-outlined";
            } else if (this.isVideoNotAllowed) {
                result += "cursor-notallowed";
            }

            return result;
        },
        videoButtonTitle() {
            // Mouse-over title text for the video button.
            let parts = ["Open video stream"];

            if (this.user.video & VideoFlag.MutualRequired) {
                parts.push("mutual video sharing required");
            }

            if (this.user.video & VideoFlag.MutualOpen) {
                parts.push("will auto-open your video");
            }

            if (this.user.video & VideoFlag.VipOnly) {
                parts.push(`${this.vipConfig.Name} only`);
            }

            if (this.user.video & VideoFlag.NonExplicit) {
                parts.push("prefers non-explicit video");
            }

            return parts.join("; ");
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
        hasReactions() {
            return this.reactions != undefined && Object.keys(this.reactions).length > 0;
        },

        // Status icons
        hasStatusIcon() {
            return this.user.status !== 'online' && this.statusMessage != undefined;
        },
        statusIconClass() {
            let status = this.statusMessage.getStatus(this.user.status);
            return status.icon;
        },
        statusLabel() {
            let status = this.statusMessage.getStatus(this.user.status);
            return `${status.emoji} ${status.label}`;
        },
    },
    methods: {
        openProfile() {
            this.$emit('open-profile', this.user.username);
        },

        // Directly open the profile page.
        openProfilePage() {
            if (this.profileURL) {
                window.open(this.profileURL);
            } else {
                this.openProfile();
            }
        },

        openDMs() {
            this.$emit('send-dm', {
                username: this.user.username,
            });
        },

        openVideo() {
            this.$emit('open-video', this.user);
        },

        muteUser() {
            this.$emit('mute-user', this.user.username);
        },

        // Boot user off your cam (for isWatchingTab)
        bootUser() {
            this.$emit('boot-user', this.user.username);
        },

        urlFor(url) {
            // Prepend the base websiteUrl if the given URL is relative.
            if (url.match(/^https?:/i)) {
                return url;
            }
            return this.websiteUrl.replace(/\/+$/, "") + url;
        },
    }
}
</script>

<template>
    <div class="columns is-mobile">
        <!-- Avatar URL if available -->
        <div class="column is-narrow pr-0" style="position: relative">
            <a :href="profileURL"
                @click.prevent="openProfile()"
                class="p-0">
                <img v-if="avatarURL" :src="avatarURL" width="24" height="24" alt="">
                <img v-else src="/static/img/shy.png" width="24" height="24">

                <!-- Away symbol -->
                <div v-if="hasStatusIcon" class="status-away-icon">
                    <i :class="statusIconClass" class="has-text-light"
                        :title="'Status: ' + statusLabel"></i>
                </div>
            </a>
        </div>
        <div class="column pr-0 is-clipped" :class="{ 'pl-1': avatarURL }">
            <strong class="truncate-text-line is-size-7 cursor-pointer"
                @click="openProfile()">
                {{ user.username }}
            </strong>
            <sup class="fa fa-peace has-text-warning is-size-7 ml-1" v-if="user.op"
                title="Operator"></sup>
            <sup class="is-size-7 ml-1" :class="vipConfig.Icon" v-else-if="user.vip"
                :title="vipConfig.Name"></sup>
        </div>
        <div class="column is-narrow pl-0">
            <!-- Emoji icon (Who's Online tab only) -->
            <span v-if="user.emoji && !isWatchingTab" class="pr-1 cursor-default" :title="user.emoji">
                {{ user.emoji.split(" ")[0] }}
            </span>

            <!-- Profile button -->
            <button type="button" class="button is-small px-2 py-1"
                :class="profileButtonClass" @click="openProfilePage()"
                :title="'Open profile page' + (user.gender ? ` (gender: ${user.gender})` : '') + (user.vip ? ` (${vipConfig.Name})` : '')">
                <i class="fa fa-user"></i>
            </button>

            <!-- Unmute User button (if muted) -->
            <button type="button" v-if="isMuted && !isBlocked" class="button is-small px-2 py-1"
                @click="muteUser()" title="This user is muted. Click to unmute them.">
                <i class="fa fa-comment-slash has-text-danger"></i>
            </button>

            <!-- DM button (if not muted) -->
            <button type="button" v-else class="button is-small px-2 py-1" @click="openDMs(u)"
                :disabled="user.username === username || (user.dnd && !isOp) || (isBlocked && !isOp)"
                :title="(user.dnd || isBlocked) ? 'This person is not accepting new DMs' : 'Send a Direct Message'">
                <i class="fa" :class="{ 'fa-comment': !(user.dnd || isBlocked), 'fa-comment-slash': user.dnd || isBlocked }"></i>
            </button>

            <!-- Video button -->
            <button type="button" class="button is-small px-2 py-1"
                :disabled="!(user.video & VideoFlag.Active)"
                :class="videoButtonClass"
                :title="videoButtonTitle"
                @click="openVideo()">
                <i class="fa" :class="videoIconClass"></i>
            </button>

            <!-- Boot from Video button (Watching tab only) -->
            <button v-if="isWatchingTab" type="button" class="button is-small px-2 py-1"
                @click="bootUser()"
                title="Kick this person off your cam">
                <i class="fa fa-user-xmark has-text-danger"></i>
            </button>
        </div>
    </div>
</template>

<style scoped>
</style>
