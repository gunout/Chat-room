<script>
export default {
    props: {
        username: String,
        message: String,
    },
    data() {
        return {
        };
    },
    computed: {
        // Message sans HTML tags, so we don't false positive on base64-encoded image data.
        filteredMessage() {
            return this.message.replace(/<(.|\n)+?>/g, "");
        },

        // Scam/spam detection and warning.
        maybeWhatsAppScam() {
            return this.filteredMessage.match(/whats\s*app/i);
        },
        maybePhoneNumberScam() {
            return this.filteredMessage.match(/\b(phone number|phone|digits|cell number|your number|ur number|text me)\b/i);
        },
        maybeOffPlatformScam() {
            return this.filteredMessage.match(/\b(telegram|signal|kik)\b/i);
        },
    },
    methods: {
    }
}
</script>

<template>
    <div v-if="maybeWhatsAppScam" class="notification is-danger is-light px-3 py-2 my-2">
        <strong class="has-text-danger">
            <i class="fa fa-exclamation-triangle mr-1"></i>
            Be careful about possible scams!
        </strong>
        It sounds like @{{ username }} is talking about moving this conversation to <i class="fab fa-whatsapp"></i> WhatsApp.
        If this happened within the first couple of messages, be wary! It is a well-known tactic for con artists to move your
        conversation away to another platform as soon as possible, in order to evade detection from the website.
        <br><br>
        <strong>Be especially skeptical of <i class="fab fa-whatsapp"></i> WhatsApp</strong> or trading phone numbers. Scammers
        can do <strong>a lot</strong> of harm with just your phone number, e.g. by plugging it into a people search website
        and bringing up lots of personal information about you.
    </div>
    <div v-else-if="maybeOffPlatformScam" class="notification is-danger is-light px-3 py-2 my-2">
        <strong class="has-text-danger">
            <i class="fa fa-exclamation-triangle mr-1"></i>
            Be careful about possible scams!
        </strong>
        It sounds like @{{ username }} is talking about moving this conversation to another messenger platform.
        If this happened within the first couple of messages, be wary! It is a well-known tactic for con artists to move your
        conversation away to another platform as soon as possible, in order to evade detection from the website.
    </div>
    <div v-else-if="maybePhoneNumberScam" class="notification is-danger is-light px-3 py-2 my-2">
        <strong class="has-text-danger">
            <i class="fa fa-exclamation-triangle mr-1"></i>
            Be careful about possible scams!
        </strong>
        It sounds like @{{ username }} may want to get your phone number. If this happened within the first couple of messages,
        be wary! Scammers can do <strong>a lot</strong> of harm with just your phone number, e.g. by plugging it into a people
        search website and bringing up lots of personal information about you.
    </div>
</template>

<style scoped>
</style>
