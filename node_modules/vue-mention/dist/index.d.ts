import { App } from 'vue';
import Mentionable from './Mentionable.vue';
export { Mentionable, };
export declare function install(app: App, options: string): void;
declare const plugin: {
    version: string;
    install: typeof install;
};
export default plugin;
