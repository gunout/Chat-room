<script>
import EmojiPicker from 'vue3-emoji-picker';
import LocalStorage from '../lib/LocalStorage';
import ScamDetection from './ScamDetection.vue';
import 'vue3-emoji-picker/css';

export default {
    props: {
        message: Object,     // chat Message object
        action: String,      // presence, notification, or (default) normal chat message
        appearance: String,  // message style appearance (cards, compact, etc.)
        user: Object,        // User object of the Message author
        isOffline: Boolean,  // user is not currently online
        username: String,    // current username logged in
        websiteUrl: String,  // Base URL to website (for profile/avatar URLs)
        isDnd: Boolean,      // user is not accepting DMs
        isMuted: Boolean,    // user is muted by current user
        reactions: Object,   // emoji reactions on the message
        reportEnabled: Boolean, // Report Message webhook is available
        position: Number,    // position of the message (0 to n), for the emoji menu to know which side to pop
        isDm: Boolean,       // is in a DM thread (hide DM buttons)
        isOp: Boolean,       // current user is Operator (always show takeback button)
        noButtons: Boolean,  // hide all message buttons (e.g. for Report Modal)
    },
    components: {
        EmojiPicker,
        ScamDetection,
    },
    data() {
        return {
            // Emoji picker visible
            showEmojiPicker: false,

            // Message menu (compact displays)
            menuVisible: false,

            // Favorite emojis
            customEmojiGroups: {
                frequently_used: [
                    { n: ['heart'], u: '2764-fe0f' },
                    { n: ['+1', 'thumbs_up'], u: '1f44d' },
                    { n: ['-1', 'thumbs_down'], u: '1f44e' },
                    { n: ['rolling_on_the_floor_laughing'], u: '1f923' },
                    { n: ['wink'], u: '1f609' },
                    { n: ['cry'], u: '1f622' },
                    { n: ['angry'], u: '1f620' },
                    { n: ['heart_eyes'], u: '1f60d' },

                    { n: ['kissing_heart'], u: '1f618' },
                    { n: ['wave'], u: '1f44b' },
                    { n: ['fire'], u: '1f525' },
                    { n: ['smiling_imp'], u: '1f608' },
                    { n: ['peach'], u: '1f351' },
                    { n: ['eggplant', 'aubergine'], u: '1f346' },
                    { n: ['splash', 'sweat_drops'], u: '1f4a6' },
                    { n: ['banana'], u: '1f34c' },
                ]
            },

            // Emoji reactions are toggled fully spelled out (for mobile)
            showReactions: false,
        };
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
        hasReactions() {
            return this.reactions != undefined && Object.keys(this.reactions).length > 0;
        },

        // Compactify a message (remove paragraph breaks added by Markdown renderer)
        compactMessage() {
            return this.message.message.replace(/<\/p>\s*<p>/g, "<br><br>").replace(/<\/?p>/g, "");
        },

        emojiPickerTheme() {
            let theme = LocalStorage.get('theme');
            if (theme === 'light' || theme === 'dark') {
                return theme;
            }
            return 'auto';
        },
    },
    methods: {
        openProfile() {
            this.$emit('open-profile', this.message.username);
        },

        openDMs() {
            this.$emit('send-dm', {
                username: this.message.username,
            });
        },

        muteUser() {
            this.$emit('mute-user', this.message.username);
        },

        takeback() {
            this.$emit('takeback', this.message);
        },

        removeMessage() {
            this.$emit('remove', this.message);
        },

        reportMessage() {
            this.$emit('report', this.message);
        },

        sendReact(emoji) {
            this.$emit('react', this.message, emoji);
        },

        // Vue3-emoji-picker callback
        onSelectEmoji(emoji) {
            this.sendReact(emoji.i);
            this.hideEmojiPicker();
        },

        // Hide the emoji menu (after sending an emoji or clicking the react button again)
        hideEmojiPicker() {
            if (!this.showEmojiPicker) return;
            window.requestAnimationFrame(() => {
                this.showEmojiPicker = false;
            });
        },

        urlFor(url) {
            // Prepend the base websiteUrl if the given URL is relative.
            if (url.match(/^https?:/i)) {
                return url;
            }
            return this.websiteUrl.replace(/\/+$/, "") + url;
        },

        // Current user has reacted to the message.
        iReacted(emoji) {
            if (!this.hasReactions) return false;

            // test whether the current user has reacted
            if (this.reactions[emoji] != undefined) {
                for (let reactor of this.reactions[emoji]) {
                    if (reactor === this.username) {
                        return true;
                    }
                }
            }
            return false;
        },

        // Google Translate link.
        translate() {
            let message = this.message?.message.replace(/<(.|\n)+?>/g, "");
            let url = `https://translate.google.com/?sl=auto&tl=en&text=${encodeURIComponent(message)}&op=translate`;
            window.open(url);
        },

        // TODO: DRY
        prettyDate(date) {
            if (date == undefined) return '';
            let hours = date.getHours(),
                minutes = String(date.getMinutes()).padStart(2, '0'),
                ampm = hours >= 12 ? "pm" : "am";

            let hour = hours % 12 || 12;
            return `${(hour)}:${minutes} ${ampm}`;
        },

        prettyDateCompact(date) {
            if (date == undefined) return '';
            let hour = date.getHours(),
                minutes = String(date.getMinutes()).padStart(2, '0');
            return `${hour}:${minutes}`;
        },
    }
}
</script>

<template>
    <!-- Presence message banners -->
    <div v-if="action === 'presence'" class="notification is-success is-light py-1 px-3 mb-2">

        <!-- Tiny avatar next to name and action buttons -->
        <div class="columns is-mobile">
            <div class="column is-narrow pr-0 pt-4">
                <a :href="profileURL" @click.prevent="openProfile()" :class="{ 'cursor-default': !profileURL }">
                    <figure class="image is-16x16">
                        <img v-if="avatarURL" :src="avatarURL">
                        <img v-else src="/static/img/shy.png">
                    </figure>
                </a>
            </div>
            <div class="column">
                <!-- Timestamp on the right -->
                <span class="float-right is-size-7" :title="message.at">
                    {{ prettyDate(message.at) }}
                </span>

                <span @click="openProfile()" class="cursor-pointer">
                    <strong>{{ nickname }}</strong>
                    <span v-if="isOffline" class="ml-1">(offline)</span>
                    <small v-else class="ml-1">(@{{ message.username }})</small>
                </span>
                {{ message.message }}
            </div>
        </div>

    </div>

    <!-- Notification message banners (e.g. DM disclaimer) -->
    <div v-else-if="action === 'notification'" class="notification is-warning is-light mb-2">
        <span v-html="message.message"></span>
    </div>

    <!-- Card Style (default) -->
    <div v-else-if="appearance === 'cards' || !appearance" class="box mb-2 px-4 pt-3 pb-1 position-relative">
        <div class="media mb-0">
            <div class="media-left">
                <a :href="profileURL" @click.prevent="openProfile()">
                    <figure class="image is-48x48">
                        <img v-if="message.isChatServer" src="/static/img/server.png">
                        <img v-else-if="message.isChatClient" src="/static/img/client.png">
                        <img v-else-if="avatarURL" :src="avatarURL">
                        <img v-else src="/static/img/shy.png">
                    </figure>
                </a>
            </div>
            <div class="media-content">
                <div class="columns is-mobile pb-0">
                    <div class="column is-narrow pb-0">
                        <strong :class="{
                            'has-text-success is-dark': message.isChatServer,
                            'has-text-warning is-dark': message.isAdmin,
                            'has-text-danger': message.isChatClient
                        }">

                            <!-- User nickname/display name -->
                            {{ nickname }}

                            <!-- Offline now? -->
                            <span v-if="isOffline">(offline)</span>
                        </strong>
                    </div>
                    <div class="column has-text-right pb-0">
                        <small class="has-text-grey is-size-7" :title="message.at">{{ prettyDate(message.at) }}</small>
                    </div>
                </div>

                <!-- User @username below it which may link to a profile URL if JWT -->
                <div class="columns is-mobile pt-0" v-if="(message.isChatClient || message.isChatServer)">
                    <div class="column is-narrow pt-0">
                        <small v-if="!(message.isChatClient || message.isChatServer)">
                            <a v-if="profileURL" :href="profileURL" target="_blank" @click.prevent="openProfile()"
                                class="has-text-grey">
                                @{{ message.username }}
                            </a>
                            <span v-else class="has-text-grey">@{{ message.username }}</span>
                        </small>
                        <small v-else class="has-text-grey">internal</small>
                    </div>
                </div>
                <div v-else class="columns is-mobile pt-0">
                    <div class="column is-narrow pt-0">
                        <small v-if="!(message.isChatClient || message.isChatServer)">
                            <a :href="profileURL || '#'" target="_blank" @click.prevent="openProfile()"
                                class="has-text-grey">
                                @{{ message.username }}
                            </a>
                        </small>
                        <small v-else class="has-text-grey">internal</small>
                    </div>

                    <div class="column is-narrow pl-1 pt-0" v-if="!noButtons">
                        <!-- DMs button -->
                        <button type="button" v-if="!(message.username === username || isDm)"
                            class="button is-small px-2" @click="openDMs()"
                            :title="isDnd ? 'This person is not accepting new DMs' : 'Open a Direct Message (DM) thread'"
                            :disabled="isDnd">
                            <i class="fa fa-comment"></i>
                        </button>

                        <!-- Mute button -->
                        <button type="button" v-if="!(message.username === username)"
                            class="button is-small px-2 ml-1" @click="muteUser()" title="Mute user">
                            <i class="fa fa-comment-slash" :class="{
                                'has-text-success': isMuted,
                                'has-text-danger': !isMuted
                            }"></i>
                        </button>

                        <!-- Owner or admin: take back the message -->
                        <button type="button" v-if="message.username === username || isOp"
                            class="button is-small px-2 ml-1"
                            title="Take back this message (delete it for everybody)" @click="takeback()"
                            :data-msgid="message.msgID">
                            <i class="fa fa-rotate-left has-text-danger"></i>
                        </button>

                        <!-- Everyone else: can hide it locally -->
                        <button type="button" v-if="message.username !== username"
                            class="button is-small px-2 ml-1"
                            title="Hide this message (delete it only for your view)" @click="removeMessage()">
                            <i class="fa fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Report & Emoji buttons -->
        <div v-if="message.msgID && !noButtons" class="emoji-button columns is-mobile is-gapless mb-0">
            <!-- Translate message button -->
            <div class="column">
                <button class="button is-small mr-1 py-2 has-text-success"
                    title="Translate this message using Google Translate"
                    @click.prevent="translate()">
                    <i class="fab fa-google has-text-success"></i>
                </button>
            </div>

            <!-- Report message button -->
            <div class="column" v-if="reportEnabled && message.username !== username">
                <button class="button is-small is-outlined mr-1 py-2" :class="{
                    'is-danger': !message.reported,
                    'has-text-grey': message.reported
                }" title="Report this message" @click.prevent="reportMessage()">
                    <i class="fa fa-flag"></i>
                    <i class="fa fa-check ml-1" v-if="message.reported"></i>
                </button>
            </div>

            <div class="column dropdown is-right"
                :class="{ 'is-up': position >= 2, 'is-active': showEmojiPicker }"
                @click="showEmojiPicker = true">
                <div class="dropdown-trigger">
                    <button type="button" class="button is-small px-2" aria-haspopup="true"
                        :aria-controls="`react-menu-${message.msgID}`" @click.prevent="hideEmojiPicker()">
                        <span>
                            <i class="fa fa-heart has-text-grey"></i>
                            <i class="fa fa-plus has-text-grey pl-1"></i>
                        </span>
                    </button>
                </div>
                <div class="dropdown-menu" :id="`react-menu-${message.msgID}`" role="menu">
                    <div class="dropdown-content p-0">
                        <!-- Emoji reactions menu -->
                        <EmojiPicker v-if="showEmojiPicker" :native="true" :display-recent="true" :disable-skin-tones="true"
                            :additional-groups="customEmojiGroups" :group-names="{ frequently_used: 'Frequently Used' }"
                            :theme="emojiPickerTheme" @select="onSelectEmoji"></EmojiPicker>
                    </div>
                </div>
            </div>
        </div>

        <!-- Message box -->
        <div class="content pl-5 pb-3 pt-1 mb-5">
            <em v-if="message.action === 'presence'">{{ message.message }}</em>
            <div v-else v-html="message.message"></div>

            <!-- Possible scam message disclaimer -->
            <ScamDetection v-if="message.username !== username"
                :username="message.username"
                :message="message.message">
            </ScamDetection>

            <!-- Reactions so far? -->
            <div v-if="hasReactions" class="mt-1">
                <span v-for="(users, emoji) in reactions" v-bind:key="emoji" class="tag mr-1 cursor-pointer"
                    :class="{ 'has-text-weight-bold': iReacted(emoji), 'is-secondary': !iReacted(emoji) }"
                    :title="emoji + ' by: ' + users.join(', ')" @click="sendReact(emoji)">
                    {{ emoji }}

                    <small v-if="showReactions" class="ml-1">
                        {{ users.join(', ') }}
                    </small>
                    <small v-else class="ml-1">{{ users.length }}</small>
                </span>

                <!-- Mobile helper to show all -->
                <a href="#" class="tag is-secondary cursor-pointer" @click.prevent="showReactions = !showReactions">
                    <i class="fa mr-1"
                        :class="{'fa-angles-left': showReactions,
                                 'fa-angles-right': !showReactions,
                        }"></i> {{ showReactions ? 'Less' : 'More' }}
                </a>
            </div>
        </div>

    </div>

    <!-- Compact styles (with or without usernames) -->
    <div v-else-if="appearance.indexOf('compact') === 0" class="columns is-mobile">
        <!-- Timestamp -->
        <div class="column is-narrow pr-0">
            <small class="has-text-grey is-size-7" :title="message.at">{{ prettyDateCompact(message.at) }}</small>
        </div>

        <!-- Avatar icon -->
        <div class="column is-narrow px-1">
            <a :href="profileURL" @click.prevent="openProfile()" class="p-0">
                <img v-if="avatarURL" :src="avatarURL" width="16" height="16" alt="">
                <img v-else src="/static/img/shy.png" width="16" height="16">
            </a>
        </div>

        <!-- Name/username/message -->
        <div class="column px-1">
            <div class="content mb-2">
                <strong :class="{
                    'has-text-success is-dark': message.isChatServer,
                    'has-text-warning is-dark': message.isAdmin,
                    'has-text-danger': message.isChatClient
                }">
                    [<a :href="profileURL" @click.prevent="openProfile()" class="has-text-dark"
                        :class="{ 'cursor-default': !profileURL }">
                        <!-- Display name? -->
                        <span v-if="(message.isChatServer || message.isChatClient || message.isAdmin)
                            || (appearance === 'compact' && nickname !== message.username)" :class="{
        'has-text-success is-dark': message.isChatServer,
        'has-text-warning is-dark': message.isAdmin,
        'has-text-danger': message.isChatClient
    }">
                            {{ nickname }}
                        </span>

                        <small class="has-text-grey"
                            :class="{ 'ml-1': appearance === 'compact' && nickname !== message.username }"
                            v-if="!(message.isChatServer || message.isChatClient || message.isAdmin)">@{{ message.username
                            }}</small>
                    </a>]
                </strong>

                <span v-html="compactMessage"></span>

                <!-- Possible scam message disclaimer -->
                <ScamDetection v-if="message.username !== username"
                    :username="message.username"
                    :message="message.message">
                </ScamDetection>
            </div>

            <!-- Reactions so far? -->
            <div v-if="hasReactions" class="mb-2">
                <span v-for="(users, emoji) in reactions" v-bind:key="emoji" class="tag mr-1 cursor-pointer"
                    :class="{ 'has-text-weight-bold': iReacted(emoji), 'is-secondary': !iReacted(emoji) }"
                    :title="emoji + ' by: ' + users.join(', ')" @click="sendReact(emoji)">
                    {{ emoji }}

                    <small v-if="showReactions" class="ml-1">
                        {{ users.join(', ') }}
                    </small>
                    <small v-else class="ml-1">{{ users.length }}</small>
                </span>

                <!-- Mobile helper to show all -->
                <a href="#" class="tag is-secondary cursor-pointer" @click.prevent="showReactions = !showReactions">
                    <i class="fa mr-1"
                        :class="{'fa-angles-left': showReactions,
                                 'fa-angles-right': !showReactions,
                        }"></i> {{ showReactions ? 'Less' : 'More' }}
                </a>
            </div>

        </div>

        <!-- Emoji/Menu button -->
        <div v-if="message.msgID && !noButtons" class="column is-narrow pl-1">

            <div class="columns is-mobile is-gapless mb-0">
                <!-- More buttons menu (DM, mute, report, etc.) -->
                <div class="column dropdown is-right"
                    :class="{ 'is-up': position >= 2, 'is-active': menuVisible }"
                    @click="menuVisible = !menuVisible">
                    <div class="dropdown-trigger">
                        <button type="button" class="button is-small px-2 mr-1" aria-haspopup="true"
                            :aria-controls="`msg-menu-${message.msgID}`">
                            <small>
                                <i class="fa fa-ellipsis-vertical"></i>
                            </small>
                        </button>
                    </div>
                    <div class="dropdown-menu" :id="`msg-menu-${message.msgID}`" role="menu">
                        <div class="dropdown-content">
                            <a href="#" class="dropdown-item" v-if="message.username !== username"
                                @click.prevent="openDMs()">
                                <i class="fa fa-comment mr-1"></i> Direct Message
                            </a>

                            <a href="#" class="dropdown-item" v-if="!(message.username === username)"
                                @click.prevent="muteUser()">
                                <i class="fa fa-comment-slash mr-1" :class="{
                                    'has-text-success': isMuted,
                                    'has-text-danger': !isMuted
                                }"></i>
                                <span v-if="isMuted">Unmute user</span>
                                <span v-else>Mute user</span>
                            </a>

                            <a href="#" class="dropdown-item" v-if="message.username === username || isOp"
                                @click.prevent="takeback()" :data-msgid="message.msgID">
                                <i class="fa fa-rotate-left has-text-danger mr-1"></i>
                                Take back
                            </a>

                            <a href="#" class="dropdown-item" v-if="message.username !== username"
                                @click.prevent="removeMessage()">
                                <i class="fa fa-trash mr-1"></i>
                                Hide message
                            </a>

                            <!-- Google Translate -->
                            <a href="#" class="dropdown-item"
                                @click.prevent="translate()">
                                <i class="fab fa-google has-text-success mr-1"></i>
                                Google Translate <i class="fa fa-external-link ml-1"></i>
                            </a>

                            <!-- Report button -->
                            <a href="#" class="dropdown-item" v-if="reportEnabled && message.username !== username"
                                @click.prevent="reportMessage()">
                                <i class="fa fa-flag mr-1" :class="{ 'has-text-danger': !message.reported }"></i>
                                <span v-if="message.reported">Reported</span>
                                <span v-else>Report</span>
                            </a>
                        </div>
                    </div>
                </div>

                <!-- Emoji reactions -->
                <div class="column dropdown is-right"
                    :class="{ 'is-up': position >= 2, 'is-active': showEmojiPicker }"
                    @click="showEmojiPicker = true">
                    <div class="dropdown-trigger">
                        <button type="button" class="button is-small px-2" aria-haspopup="true"
                            :aria-controls="`react-menu-${message.msgID}`" @click="hideEmojiPicker()">
                            <small>
                                <i class="fa fa-heart has-text-grey"></i>
                            </small>
                        </button>
                    </div>
                    <div class="dropdown-menu" :id="`react-menu-${message.msgID}`" role="menu">
                        <div class="dropdown-content p-0">
                            <!-- Emoji reactions menu -->
                            <EmojiPicker v-if="showEmojiPicker" :native="true" :display-recent="true"
                                :disable-skin-tones="true" :additional-groups="customEmojiGroups"
                                :group-names="{ frequently_used: 'Frequently Used' }" theme="auto" @select="onSelectEmoji">
                            </EmojiPicker>
                        </div>
                    </div>
                </div>
            </div>
        </div>

    </div>
</template>

<style scoped></style>
