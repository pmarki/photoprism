<template>
  <p-navigation>
    <template v-slot:title>
      <span class="text-capitalize">{{ this.$route.name }}</span>
    </template>

    <template v-slot:menu-icons>
      <span>
      <a
        href="#"
        :title="$gettext('Mosaic')"
        class="menu-action"
        @click.prevent="setView('mosaic')"
      >
        <v-icon>mdi-view-comfy</v-icon>
      </a>
      <a
        href="#"
        :title="$gettext('Cards')"
        class="menu-action"
        @click.prevent="setView('cards')"
      >
        <v-icon>mdi-view-column</v-icon>
      </a>
      </span>
    </template>
    <template v-slot:menu-actions>
      <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
    </template>
  </p-navigation>
</template>
<script>
import * as options from "options/options";
import $api from "common/api";
import $notify from "common/notify";
import links from "common/links";

import PActionMenu from "component/action/menu.vue";
import PConfirmDialog from "component/confirm/dialog.vue";
import PNavigation from "../navigation.vue";

export default {
  name: "PPhotoToolbar",
  components: {
    PNavigation,
    PActionMenu,
    PConfirmDialog,
  },
  props: {
    context: {
      type: String,
      default: "photos",
    },
    filter: {
      type: Object,
      default: () => {},
    },
    staticFilter: {
      type: Object,
      default: () => {},
    },
    updateFilter: {
      type: Function,
      default: () => {},
    },
    updateQuery: {
      type: Function,
      default: () => {},
    },
    settings: {
      type: Object,
      default: () => {},
    },
    refresh: {
      type: Function,
      default: () => {},
    },
    onClose: {
      type: Function,
      default: undefined,
    },
    embedded: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const features = this.$config.getSettings().features;
    const readonly = this.$config.get("readonly");

    return {
      expanded: false,
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      isSuperAdmin: this.$session.isSuperAdmin(),
      config: this.$config.values,
      readonly: readonly,
      canUpload: !readonly && !this.embedded && this.$config.allow("files", "upload") && features.upload,
      canDelete: !readonly && !this.embedded && this.$config.allow("photos", "delete") && features.delete,
      canAccessLibrary: this.$config.allow("photos", "access_library"),
      featSettings: features.settings,
      listView: this.$config.getSettings()?.search?.listView,
      all: {
        countries: [{ ID: "", Name: this.$gettext("All Countries") }],
        cameras: [{ ID: 0, Name: this.$gettext("All Cameras") }],
        lenses: [{ ID: 0, Name: this.$gettext("All Lenses") }],
        colors: [{ Slug: "", Name: this.$gettext("All Colors") }],
        categories: [{ Slug: "", Name: this.$gettext("All Categories") }],
        months: [{ value: 0, text: this.$gettext("All Months") }],
        years: [{ value: 0, text: this.$gettext("All Years") }],
      },
      dialog: {
        delete: false,
      },
    };
  },
  computed: {
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
    countryOptions() {
      return this.all.countries.concat(this.config.countries);
    },
    cameraOptions() {
      return this.all.cameras.concat(this.config.cameras);
    },
    categoryOptions() {
      return this.all.categories.concat(this.config.categories);
    },
    viewOptions() {
      if (this.$config.getSettings()?.search?.listView) {
        return [
          { value: "mosaic", text: this.$gettext("Mosaic") },
          { value: "cards", text: this.$gettext("Cards") },
          { value: "list", text: this.$gettext("List") },
        ];
      } else {
        return [
          { value: "mosaic", text: this.$gettext("Mosaic") },
          { value: "cards", text: this.$gettext("Cards") },
        ];
      }
    },
    sortOptions() {
      switch (this.context) {
        case "archive":
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "archived", text: this.$gettext("Recently Archived") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
          ];
        case "hidden":
        case "review":
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
          ];
        default:
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "edited", text: this.$gettext("Recently Edited") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
            { value: "similar", text: this.$gettext("Visual Similarity") },
            { value: "relevance", text: this.$gettext("Most Relevant") },
          ];
      }
    },
  },
  methods: {
    showExpansionPanel() {
      if (!this.expanded) {
        this.expanded = true;
      }
    },
    hideExpansionPanel() {
      if (this.expanded) {
        this.expanded = false;
      }
    },
    toggleExpansionPanel() {
      this.expanded = !this.expanded;
    },
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload"),
          shortcut: "Ctrl-U",
          visible: this.canUpload && this.context !== "archive" && this.context !== "hidden",
          click: () => {
            this.showUpload();
          },
        },
      ];
    },
    colorOptions() {
      return this.all.colors.concat(options.Colors());
    },
    monthOptions() {
      return this.all.months.concat(options.Months());
    },
    yearOptions() {
      return this.all.years.concat(options.IndexedYears());
    },
    setView(name) {
      if (name) {
        if (name === "list" && !this.listView) {
          name = "mosaic";
        }
        this.hideExpansionPanel();
        this.refresh({ view: name });
      }
    },
    showUpload() {
      this.$event.publish("dialog.upload");
    },
    deleteAll() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = true;
    },
    clearLocation() {
      this.$router.push({ name: "browse" });
    },
    onBrowse() {
      const route = { name: "places_browse", query: this.staticFilter };

      if (this.$isMobile) {
        this.$router.push(route);
      } else {
        // Open in a new tab on desktop browsers.
        const routeUrl = this.$router.resolve(route).href;

        if (routeUrl) {
          window.open(routeUrl, "_blank");
        }
      }
    },
    onUpdate(v) {
      this.updateQuery(v);
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      $api.post("batch/photos/delete", { all: true }).then(() => this.onDeleted());
    },
    onDeleted() {
      $notify.success(this.$gettext("Permanently deleted"));
      this.$clipboard.clear();
    },
  },
};
</script>
