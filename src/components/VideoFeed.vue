<script>
import Slider from 'vue3-slider';

export default {
    props: {
        localVideo: Boolean,  // is our local webcam (not other's camera)
        poppedOut: Boolean,   // Video is popped-out and draggable
        username: String,     // username related to this video
        isExplicit: Boolean,  // camera is marked Explicit
        isMuted: Boolean,     // camera is muted on our end
        isSourceMuted: Boolean, // camera is muted on the broadcaster's end
        isWatchingMe: Boolean, // other video is watching us back
        isFrozen: Boolean,     // video is detected as frozen
        isSpeaking: Boolean,   // video is registering audio
        watermarkImage: Image, // watermark image to overlay (nullable)
    },
    components: {
        Slider,
    },
    data() {
        return {
            // Volume slider
            volume: 100,

            // Volume change debounce
            volumeDebounce: null,

            // Mouse over status
            mouseOver: false,
        };
    },
    computed: {
        containerID() {
            return this.videoID + '-container';
        },
        videoID() {
            return this.localVideo ? 'localVideo' : `videofeed-${this.username}`;
        },
        textColorClass() {
            return this.isExplicit ? 'has-text-camera-red' : 'has-text-camera-blue';
        },
        muteButtonClass() {
            let classList = [
                'button is-small ml-1 px-2',
            ]

            if (this.isMuted) {
                classList.push('is-danger');
            } else {
                classList.push('is-success is-outlined');
            }
            return classList.join(' ');
        },
        muteIconClass() {
            if (this.localVideo) {
                return this.isMuted ? 'fa-microphone-slash' : 'fa-microphone';
            }
            return this.isMuted ? 'fa-volume-xmark' : 'fa-volume-high';
        }
    },
    methods: {
        closeVideo() {
            // Note: closeVideo only available for OTHER peoples cameras.
            // Closes the WebRTC connection as the offerer.
            this.$emit('close-video', this.username, 'offerer');
        },

        reopenVideo() {
            // Note: goes into openVideo(username, force)
            this.$emit('reopen-video', this.username, true);
        },

        openProfile() {
            this.$emit('open-profile', this.username);
        },

        // Toggle the Mute button
        muteVideo() {
            this.$emit('mute-video', this.username);
        },

        popoutVideo() {
            this.$emit('popout', this.username);
        },

        fullscreen(force=false) {
            // If we are popped-out, pop back in before full screen.
            if (this.poppedOut && !force) {
                this.popoutVideo();
                window.requestAnimationFrame(() => {
                    this.fullscreen(true);
                });
                return;
            }

            let $elem = document.getElementById(this.containerID);
            if ($elem) {
                if (document.fullscreenElement) {
                    document.exitFullscreen();
                } else if ($elem.requestFullscreen) {
                    $elem.requestFullscreen();
                } else {
                    window.alert("Fullscreen not supported by your browser.");
                }
            }
        },

        volumeChanged() {
            if (this.volumeDebounce !== null) {
                clearTimeout(this.volumeDebounce);
            }
            this.volumeDebounce = setTimeout(() => {
                this.$emit('set-volume', this.username, this.volume);
            }, 200);
        },
    }
}
</script>

<template>
    <div class="feed" :id="containerID" :class="{
        'popped-out': poppedOut,
        'popped-in': !poppedOut,
    }" @mouseover="mouseOver = true" @mouseleave="mouseOver = false">
        <video
            :id="videoID"
            autoplay
            disablepictureinpicture
            playsinline
            oncontextmenu="return false;"
            :muted="localVideo"></video>

        <!-- Watermark layer -->
        <div v-if="watermarkImage">
            <img :src="watermarkImage" class="watermark">
            <img :src="watermarkImage" class="corner-watermark seethru invert-color">
        </div>

        <!-- Caption -->
        <div class="caption" :class="textColorClass">
            <i class="fa fa-microphone-slash mr-1 has-text-grey" v-if="isSourceMuted"></i>
            <a href="#" @click.prevent="openProfile" :class="textColorClass">{{ username }}</a>
            <i class="fa fa-people-arrows ml-1 has-text-grey is-size-7" :title="username + ' is watching your camera too'"
                v-if="isWatchingMe"></i>

            <!-- Frozen stream detection -->
            <a class="fa fa-mountain ml-1" href="#" v-if="!localVideo && isFrozen" style="color: #00FFFF"
                @click.prevent="reopenVideo()" title="Frozen video detected!"></a>

            <!-- Is speaking -->
            <span v-if="isSpeaking" class="ml-1" title="Speaking">
                <i class="fa fa-volume-high has-text-info"></i>
            </span>
        </div>

        <!-- Close button (others' videos only) -->
        <div class="close" v-if="!localVideo" :class="{'seethru': !mouseOver}">
            <a href="#" class="button is-small is-danger is-outlined px-2" title="Close video" @click.prevent="closeVideo()">
                <i class="fa fa-close"></i>
            </a>
        </div>

        <!-- Controls -->
        <div class="controls">
            <!-- Mute Button -->
            <button type="button" :class="muteButtonClass"
                @click="muteVideo()">
                <i class="fa" :class="muteIconClass"></i>
            </button>

            <!-- Pop-out Video -->
            <button type="button" class="button is-small is-light is-outlined p-2 ml-2" title="Pop out"
                :class="{'seethru': !mouseOver}"
                @click="popoutVideo()">
                <i class="fa fa-up-right-from-square"></i>
            </button>

            <!-- Full screen. -->
            <button type="button" class="button is-small is-light is-outlined p-2 ml-2" title="Go full screen"
                :class="{'seethru': !mouseOver}"
                @click="fullscreen()">
                <i class="fa fa-expand"></i>
            </button>
        </div>

        <!-- Volume slider -->
        <div class="volume-slider" v-show="!localVideo && !isMuted && mouseOver">
            <Slider v-model="volume" color="#00FF00" track-color="#006600" :min="0" :max="100" :step="1" :height="7"
                orientation="vertical" @change="volumeChanged">

            </Slider>
        </div>
    </div>
</template>

<style scoped>
.volume-slider {
    position: absolute;
    left: 18px;
    top: 30px;
    bottom: 44px;
}

/* A background image behind video elements in case they don't load properly */
video {
    background-image: url(/static/img/connection-error.png);
    background-position: center center;
    background-repeat: no-repeat;
}

/* Translucent controls until mouse over */
.seethru {
    opacity: 0.4;
}

/* Watermark image */
.watermark {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    margin: auto;
    width: 40%;
    height: 40%;
    opacity: 0.02;
    animation-name: subtle-pulsate;
    animation-duration: 10s;
    animation-iteration-count: infinite;
}
.corner-watermark {
    position: absolute;
    right: 4px;
    bottom: 4px;
    width: 20%;
    min-width: 32px;
    min-height: 32px;
    max-height: 20%;
}
.invert-color {
    filter: invert(100%);
}

/* Animate the primary watermark to pulsate in opacity */
@keyframes subtle-pulsate {
    0% { opacity: 0.02; }
    50% { opacity: 0.04; }
    100% { opacity: 0.02; }
}
</style>
