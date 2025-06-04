<template>
  <v-dialog
    :model-value="visible"
    :fullscreen="$vuetify.display.mdAndDown"
    scrim
    scrollable
    persistent
    class="p-move-dialog v-dialog--move"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
    @keydown.esc.exact="onClose"
    @focusout="onFocusOut"
  >
    <v-form ref="form" class="p-photo-move" validate-on="invalid-input" tabindex="1" @submit.prevent="onSubmit">
      <v-card :tile="$vuetify.display.mdAndDown">
        <v-toolbar
          v-if="$vuetify.display.mdAndDown"
          flat
          color="navigation"
          class="mb-4"
          :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
        >
          <v-btn icon @click.stop="onClose">
            <v-icon>mdi-close</v-icon>
          </v-btn>
          <v-toolbar-title>
            {{ title }}
          </v-toolbar-title>
        </v-toolbar>
        <v-card-title v-else class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-folder-move</v-icon>
          <h6 class="text-h6">{{ title }}</h6>
        </v-card-title>
        <v-card-text class="flex-grow-0">
          <div class="form-container">
            <div class="form-header">
              <span>{{ $gettext(`Move files to selected folder`) }}</span>
            </div>
            <div class="form-body">
              <div class="form-controls">
                <v-select
                  v-model="destination"
                  :disabled="busy || loading"
                  hide-details
                  class="input-albums"
                  density="comfortable"
                  :items="folders"
                  item-title="Path"
                  item-value="UID"
                  return-object
                >
                </v-select>
              </div>
            </div>
          </div>
        </v-card-text>
        <v-card-actions class="action-buttons mt-1">
          <v-btn :disabled="busy" variant="flat" color="button" class="action-close" @click.stop="onClose">
            {{ $gettext(`Close`) }}
          </v-btn>
          <v-btn
            :disabled="busy"
            variant="flat"
            color="highlight"
            class="action-select action-move"
            @click.stop="moveFiles()"
          >
            {{ $gettext(`Move`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import $api from "common/api";
import $notify from "common/notify";
import Album from "model/album";

export default {
  name: "PMoveDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    selection: {
      type: Array,
      default: () => [],
    },
  },
  emits: ["close", "confirm"],
  data() {
    const isDemo = this.$config.get("demo");
    return {
      accept: this.$config.get("uploadAllow"),

      folders: [],
      destination: null,
      busy: false,
      loading: false,
      failed: false,
      token: "",
      isDemo: isDemo,
    };
  },
  computed: {
    title() {
      return this.$gettext(`Move files`);
    },
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();
        this.isDemo = this.$config.get("demo");

        // Fetch folders from backend.
        this.load("");
      } else {
        this.reset();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    onFocusOut(ev) {
      if (!this.$view.isActive(this)) {
        return;
      }

      if (ev.target && ev.target instanceof HTMLElement && this.$refs.form?.$el instanceof HTMLElement) {
        if (
          document.activeElement !== this.$refs.form.$el &&
          (!ev.target.closest(".p-move-dialog") || ev.target?.disabled)
        ) {
          this.$refs.form?.$el.focus();
        }
      }
    },
    onLoad() {
      this.loading = true;
    },
    onLoaded() {
      this.loading = false;
    },
    load(q) {
      if (this.loading) {
        return;
      }

      this.onLoad();
      const params = {
        q: q,
        count: 2000,
        offset: 0,
        type: "folder",
        order: "name",
      };

      Album.search(params)
        .then((response) => {
          this.folders = response.models;
          console.log('album ' +  this.$route.params.album);
          this.folders.forEach((item) => {
            if (item.UID === this.$route.params.album) {
              this.destination = item;
            }
          });
        })
        .finally(() => {
          this.onLoaded();
        });
    },
    onClose() {
      if (this.busy) {
        $notify.info(this.$gettext("Moving photos…"));
        return;
      }

      this.$emit("close");
    },
    confirm() {
      if (this.busy) {
        $notify.info(this.$gettext("Moving photos…"));
        return;
      }

      this.$emit("confirm");
    },
    onSubmit() {
      // DO NOTHING
    },
    reset() {
      this.busy = false;
      this.destination= null;
      this.token = "";
    },
    moveFiles() {
      if (this.busy) {
        return;
      }
      let data = [];

      this.selection.forEach((item) => data.push({PhotoUID: item, Dest: this.destination.Path}));

      $api
        .put(`files/move`, {
          Files: data,
        })
        .then(() => {
          this.$notify.success(this.$gettext("Done"));
          this.$emit("confirm");
        })
        .catch(() => {
          this.reset();
          $notify.error(this.$gettext("Moving files failed"));
        });

    },
  },
};
</script>
