<template>
  <div ref="page" tabindex="1" class="p-page p-page-albums not-selectable" :class="$config.aclClasses('albums')">
    <p-navigation>
      <template v-slot:title>
        <span class="text-capitalize">{{ this.$route.name }}</span>
      </template>

      <template v-slot:menu-icons>

      </template>
      <template v-slot:menu-actions>
        <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
      </template>
    </p-navigation>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-scroll
        :hide-panel="hideExpansionPanel"
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      >
      </p-scroll>

      <p-album-clipboard
        :refresh="refresh"
        :selection="selection"
        :share="share"
        :edit="edit"
        :clear-selection="clearSelection"
        :context="context"
      ></p-album-clipboard>

      <div v-if="results.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-lightbulb-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">
            {{ $gettext(`No albums found`) }}
          </div>
        </v-alert>
      </div>
      <div v-else>
        <div v-for="years in byYear" :key="years.year">
          <h4 class="text-h5 mt-3 pl-3">{{ years.year }}</h4>
          <div
            class="v-row search-results album-results cards-view"
            :class="{ 'select-results': selection.length > 0 }"
          >
            <div
              v-for="album in years.albums"
              :key="album.UID"
              ref="items"
              class="v-col-6 v-col-sm-4 v-col-md-3 v-col-xl-2"
            >
          <div
            :data-uid="album.UID"
            class="result not-selectable"
            :class="album.classes(selection.includes(album.UID))"
            @contextmenu.stop="onContextMenu($event, album.Index)"
          >
            <div
              :key="album.UID"
              :title="album.Title"
              class="preview"
              @touchstart.passive="input.touchStart($event, album.Index)"
              @touchend.stop="onClick($event, album.Index)"
              @mousedown.stop.prevent="input.mouseDown($event, album.Index)"
              @click.stop.prevent="onClick($event, album.Index)"
            >
              <div class="position-relative w-100 h-100">
                <img v-for="(tile, i) in album.getThumbs('tile_500')"
                     :key="i"
                     :src="tile"
                     class="position-absolute"
                     :class="'tile' + i"
                     alt="abc"/>
              </div>
              <div class="preview__overlay"></div>
              <button
                v-if="canShare && album.LinkCount > 0"
                class="action-share"
                @touchstart.stop="input.touchStart($event, album.Index)"
                @touchend.stop="onShare($event, album.Index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onShare($event, album.Index)"
              >
                <i class="mdi mdi-share-variant" />
              </button>
              <button
                class="input-select"
                @touchstart.stop="input.touchStart($event, album.Index)"
                @touchend.stop="onSelect($event, album.Index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onSelect($event, album.Index)"
              >
                <i class="mdi mdi-check-circle select-on" />
                <i class="mdi mdi-circle-outline select-off" />
              </button>
              <button
                v-if="canManage && experimental && featPrivate && album.Private"
                class="input-private"
                @touchstart.stop="input.touchStart($event, album.Index)"
                @touchend.stop="onEdit($event, album.Index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onEdit($event, album.Index)"
              >
                <i class="mdi mdi-lock" />
              </button>
            </div>

            <div class="meta">
              <button
                v-if="album.Type === 'month'"
                :title="album.Title"
                class="action-title-edit meta-title text-capitalize"
                :data-uid="album.UID"
              >
                {{ album.getDateString() }}
              </button>
              <button
                v-else-if="album.Type === 'folder'"
                :title="album.Title"
                class="action-title-edit meta-title"
                :data-uid="album.UID"
              >
                {{ album.Path }}
              </button>
              <button
                v-else-if="album.Title"
                :title="album.Title"
                class="action-title-edit meta-title"
                :data-uid="album.UID"
              >
                {{ album.Title }}
              </button>
            </div>
          </div>
        </div>
            </div>
        </div>
      </div>
    </div>
    <p-share-dialog
      :visible="dialog.share"
      :model="model"
      @upload="webdavUpload"
      @close="dialog.share = false"
    ></p-share-dialog>
    <p-service-upload
      :visible="dialog.upload"
      :items="{ albums: selection }"
      :model="model"
      @close="dialog.upload = false"
      @confirm="dialog.upload = false"
    ></p-service-upload>
    <p-album-edit-dialog :visible="dialog.edit" :album="model" @close="dialog.edit = false"></p-album-edit-dialog>
    <p-album-create-dialog :visible="dialog.create" @close="dialog.create = false"></p-album-create-dialog>
  </div>
</template>

<script>
import Album from "model/album";
import { DateTime } from "luxon";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import $notify from "common/notify";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import * as options from "options/options";

import PLoading from "component/loading.vue";
import PActionMenu from "component/action/menu.vue";
import PNavigation from "../component/navigation.vue";
import Folder from "../model/folder";
import PAlbumCreateDialog from "../component/album/create/dialog.vue";

export default {
  name: "PPageCalendar",
  components: {
    PAlbumCreateDialog,
    PNavigation,
    PLoading,
    PActionMenu,
  },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
    defaultOrder: {
      type: String,
      default: "name",
    },
    view: {
      type: String,
      default: "",
    },
  },
  expose: ["onShortCut"],
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const order = this.sortOrder();
    const q = query["q"] ? query["q"] : "";
    const category = query["category"] ? query["category"] : "";
    const year = query["year"] ? parseInt(query["year"]) : "";
    const filter = { q, category, order, year };
    const settings = {};
    const features = this.$config.getSettings().features;

    let categories = [{ value: "", text: this.$gettext("All Categories") }];

    if (this.$config.albumCategories().length > 0) {
      categories = categories.concat(
        this.$config.albumCategories().map((cat) => {
          return { value: cat, text: cat };
        })
      );
    }

    return {
      expanded: false,
      experimental: this.$config.get("experimental") && !this.$config.ce(),
      canUpload: this.$config.allow("files", "upload") && features.upload,
      canCreate: this.$config.allow("files", "upload") && features.upload,
      canShare: this.$config.allow("albums", "share") && features.share,
      canManage: this.$config.allow("albums", "manage"),
      canEdit: this.$config.allow("albums", "update"),
      config: this.$config.values,
      isSuperAdmin: this.$session.isSuperAdmin(),
      featShare: features.share,
      featPrivate: features.private,
      featSettings: features.settings,
      categories: categories,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      byYear: [],
      loading: true,
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 2,
      batchSize: Album.batchSize(),
      offset: 0,
      page: 0,
      selection: [],
      settings: settings,
      q: q,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Title too long"),
      input: new Input(),
      lastId: "",
      dialog: {
        share: false,
        upload: false,
        edit: false,
        create: false,
      },
      model: new Album(false),
      all: {
        years: [{ value: "", text: this.$gettext("All Years") }],
      },
      options: {
        sorting: [
          { value: "favorites", text: this.$gettext("Favorites") },
          { value: "name", text: this.$gettext("Name") },
          { value: "place", text: this.$gettext("Location") },
          { value: "newest", text: this.$gettext("Newest First") },
          { value: "oldest", text: this.$gettext("Oldest First") },
          { value: "added", text: this.$gettext("Recently Added") },
          { value: "edited", text: this.$gettext("Recently Edited") },
        ],
      },
    };
  },
  computed: {
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
    context: function () {
      if (!this.staticFilter) {
        return "album";
      }

      if (this.staticFilter.type) {
        return this.staticFilter.type;
      }

      return "";
    },
    canExpand: function () {
      return this.canManage && !this.staticFilter["order"];
    },
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      const query = this.$route.query;

      this.routeName = this.$route.name;
      this.lastFilter = {};
      this.q = query["q"] ? query["q"] : "";
      this.filter.q = this.q;
      this.filter.category = query["category"] ? query["category"] : "";
      this.filter.year = query["year"] ? parseInt(query["year"]) : "";
      this.filter.order = this.sortOrder();

      this.search();
    },
  },
  created() {
    this.search();

    this.subscriptions.push(this.$event.subscribe("albums", (ev, data) => this.onUpdate(ev, data)));
    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
    this.subscriptions.push(this.$event.subscribe("config.updated", (ev, data) => this.onConfigUpdated(data)));
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    createNew() {
      this.dialog.create = true;
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
          visible: this.canUpload,
          click: () => {
            this.showUpload();
          },
        },
        {
          name: "Create",
          icon: "mdi-plus",
          text: this.$gettext("Add New"),
          visible: this.canCreate,
          click: () => {
            this.createNew();
          },
        },
      ];
    },
    onShortCut(ev) {
      switch (ev.code) {
        case "Escape":
          if (document.activeElement instanceof HTMLInputElement) {
            document.activeElement.blur();
          }
          this.hideExpansionPanel();
          return true;
        case "KeyR":
          this.refresh();
          return true;
        case "KeyF":
          if (ev.shiftKey) {
            this.showExpansionPanel();
          }
          this.$view.focus(this.$refs?.form, ".input-search input", true);
          return true;
        case "KeyU":
          if (this.$config.allow("files", "upload") && this.$config.feature("upload")) {
            this.$event.publish("dialog.upload");
          }
          return true;
      }
    },
    toggleExpansionPanel() {
      if (!this.canExpand) {
        return;
      }

      this.expanded = !this.expanded;
    },
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
    onConfigUpdated(data) {
      if (!data || !data.config?.albumCategories) {
        return;
      }

      const c = data.config.albumCategories;

      this.categories = [{ value: "", text: this.$gettext("All Categories") }];

      if (c.length > 0) {
        this.categories = this.categories.concat(
          c.map((cat) => {
            return { value: cat, text: cat };
          })
        );
      }
    },
    yearOptions() {
      return this.all.years.concat(options.IndexedYears());
    },
    sortOrder() {
      const typeName = this.staticFilter?.type;
      const keyName = "albums.order." + typeName;
      const queryParam = this.$route.query["order"];
      const storedType = window.localStorage.getItem(keyName);

      if (queryParam) {
        window.localStorage.setItem(keyName, queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      }

      return this.defaultOrder;
    },
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("albums.offset"));

      if (this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("albums.offset", offset);
    },
    share(album) {
      if (!album || !this.canShare) {
        return;
      }

      this.model = album;
      this.dialog.share = true;
    },
    edit(album) {
      if (!album) {
        return;
      } else if (!this.canManage) {
        this.$router.push(album.route(this.view));
        return;
      }

      this.model = album;
      this.dialog.edit = true;
    },
    webdavUpload() {
      if (!this.canShare) {
        return;
      }

      this.dialog.share = false;
      this.dialog.upload = true;
    },
    showUpload() {
      if (!this.canUpload) {
        return;
      }

      // Pre-select manually managed album in upload dialog.
      if (this.context === "album" && this.selection && this.selection.length === 1) {
        return this.model
          .find(this.selection[0])
          .then((m) => this.$event.publish("dialog.upload", { albums: [m] }))
          .catch(() => this.$event.publish("dialog.upload", { albums: [] }));
      } else {
        this.$event.publish("dialog.upload", { albums: [] });
      }
    },
    toggleLike(ev, index) {
      if (!this.canManage) {
        return;
      }

      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const album = this.results[index];

      if (!album) {
        return;
      }

      album.toggleLike();
    },
    selectRange(rangeEnd, models) {
      if (!models || !models[rangeEnd] || !(models[rangeEnd] instanceof RestModel)) {
        console.warn("selectRange() - invalid arguments:", rangeEnd, models);
        return;
      }

      let rangeStart = models.findIndex((m) => m.getId() === this.lastId);

      if (rangeStart === -1) {
        this.toggleSelection(models[rangeEnd].getId());
        return 1;
      }

      if (rangeStart > rangeEnd) {
        const newEnd = rangeStart;
        rangeStart = rangeEnd;
        rangeEnd = newEnd;
      }

      for (let i = rangeStart; i <= rangeEnd; i++) {
        this.addSelection(models[i].getId());
      }

      return rangeEnd - rangeStart + 1;
    },
    onEdit(ev, index) {
      if (!this.canManage) {
        return;
      }

      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      return this.edit(this.results[index]);
    },
    onShare(ev, index) {
      if (!this.canShare) {
        return;
      }

      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      return this.share(this.results[index]);
    },
    onSelect(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      if (ev.shiftKey) {
        this.selectRange(index, this.results);
      } else {
        this.toggleSelection(this.results[index].getId());
      }
    },
    onClick(ev, index) {
      console.log(index);
      const inputType = this.input.eval(ev, index);
      const longClick = inputType === ClickLong;

      if (inputType === InputInvalid) {
        return;
      }

      if (longClick || this.selection.length > 0) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index, this.results);
        } else {
          this.toggleSelection(this.results[index].getId());
        }
      } else {
        this.$router.push(this.results[index].route(this.view));
      }
    },
    onContextMenu(ev, index) {
      if (this.$isMobile) {
        ev.preventDefault();
        ev.stopPropagation();

        if (this.results[index]) {
          this.selectRange(index, this.results);
        }
      }
    },
    loadMore() {
      if (this.scrollDisabled) return;

      this.scrollDisabled = true;
      this.listen = false;

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;

      const params = {
        count: count,
        offset: offset,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      Album.search(params)
        .then((resp) => {
          this.results = this.dirty ? resp.models : this.results.concat(resp.models);

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            this.setOffset(resp.offset);

            if (this.results.length > 1) {
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("All %{n} albums loaded"), { n: this.results.length })
              );
            }
          } else {
            this.setOffset(resp.offset + resp.limit);
            this.page++;

            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }
        })
        .catch(() => {
          this.scrollDisabled = false;
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
        });
    },
    updateSettings(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.settings.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.settings[key] = value.trim();
            break;
          default:
            this.settings[key] = value;
        }

        window.localStorage.setItem("albums." + key, this.settings[key]);
      }
    },
    updateFilter(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.filter.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.filter[key] = value.trim();
            break;
          default:
            this.filter[key] = value;
        }
      }
    },
    updateQuery(props) {
      this.updateFilter(props);

      if (this.loading) return;

      const query = {
        view: this.settings.view,
      };

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return;
      }

      this.$router.replace({ query: query });
    },
    searchParams() {
      const params = {
        count: 10000,
        offset: this.offset,
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    reset() {
      this.results = [];
    },
    search() {
      /**
       * re-creating the last scroll-position should only ever happen when using
       * back-navigation. We therefore reset the remembered scroll-position
       * in any other scenario
       */
      if (!window.backwardsNavigationDetected) {
        this.setOffset(0);
      }

      this.scrollDisabled = true;

      // Don't query the same data more than once
      if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        // this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;

      const params = this.searchParams();

      Album.search(params)
        .then((resp) => {
          // Hide search toolbar expansion panel when matching albums were found.
          if (this.offset === 0 && resp.count > 0) {
            this.hideExpansionPanel();
          }

          this.offset = resp.limit;
          console.log('searching');
          console.log(resp);

          this.results = resp.models;

          let groupedByYear = resp.models.reduce((acc, current, index) => {
            console.log(current.Year);
            const year = current.Year;
            if (!acc[year]) {
              acc[year] = [];
            }
            current.Index = index;
            acc[year].unshift(current);
            return acc;
          }, {});

          console.log(groupedByYear);
          let res = [];

          for (let [index, item] of Object.entries(groupedByYear)) {
            console.log(item);
            const row = { year: index, albums: item };
            res.unshift(row);
          };

          this.byYear = res;

          console.log(this.byYear);

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            if (!this.results.length) {
              this.$notify.warn(this.$gettext("No albums found"));
            }
          } else {
            // this.$notify.info(this.$gettext('More than 20 albums found'));
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }
        })
        .catch((err) => {
          console.log(err);
          this.$notify.error(err.message);
          this.reset();
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
        });
    },
    refresh(props) {
      this.updateSettings(props);

      // Do not refresh results if the view is already loading
      // or should not be listening for events.
      if (this.loading || !this.listen) {
        return;
      }

      // Make sure enough results are loaded to maintain the scroll position.
      if (this.page > 2) {
        this.page = this.page - 1;
      } else {
        this.page = 1;
      }

      // Flag results as dirty to force a refresh.
      this.dirty = true;

      // Enable infinite scrolling if it was disabled.
      this.scrollDisabled = false;

      this.loadMore();
    },
    create() {
      // Use month and year as default title.
      let title = DateTime.local().toFormat("LLLL yyyy");

      // Add suffix if the album title already exists.
      if (this.results.findIndex((a) => a.Title.startsWith(title)) !== -1) {
        const re = new RegExp(`${title} \\((\\d?)\\)`, "i");
        let i = 1;
        this.results.forEach((a) => {
          const found = a.Title.match(re);
          if (found && found.length > 0 && found[1]) {
            const n = parseInt(found[1]);
            if (n > i) {
              i = n;
            }
          }
        });

        title = `${title} (${i + 1})`;
      }

      const album = new Album({ Title: title, Favorite: false });

      album.save().then(() => this.$notify.success(this.$gettext("Album created")));
    },
    onSave(album) {
      album.update().then(() => this.$notify.success(this.$gettext("Changes successfully saved")));
    },
    addSelection(uid) {
      const pos = this.selection.indexOf(uid);

      if (pos === -1) {
        if (this.selection.length >= MaxItems) {
          $notify.warn(this.$gettext("Can't select more items"));
          return;
        }

        this.selection.push(uid);
        this.lastId = uid;
      }
    },
    toggleSelection(uid) {
      const pos = this.selection.indexOf(uid);

      if (pos !== -1) {
        this.selection.splice(pos, 1);
        this.lastId = "";
      } else {
        if (this.selection.length >= MaxItems) {
          $notify.warn(this.$gettext("Can't select more items"));
          return;
        }

        this.selection.push(uid);
        this.lastId = uid;
      }
    },
    removeSelection(uid) {
      const pos = this.selection.indexOf(uid);

      if (pos !== -1) {
        this.selection.splice(pos, 1);
        this.lastId = "";
      }
    },
    clearSelection() {
      this.selection.splice(0, this.selection.length);
      this.lastId = "";
    },
    onUpdate(ev, data) {
      if (!this.listen) {
        console.log("albums.onUpdate currently not listening", ev, data);
        return;
      } else if (!data || !data.entities || !Array.isArray(data.entities)) {
        console.log("albums.onUpdate received empty data", ev, data);
        return;
      }

      const type = ev.split(".")[1];

      switch (type) {
        case "updated":
          for (let i = 0; i < data.entities.length; i++) {
            const values = data.entities[i];
            const model = this.results.find((m) => m.UID === values.UID);

            if (model) {
              for (let key in values) {
                if (values.hasOwnProperty(key) && values[key] != null && typeof values[key] !== "object") {
                  model[key] = values[key];
                }
              }
            }
          }

          let categories = [{ value: "", text: this.$gettext("All Categories") }];

          if (this.$config.albumCategories().length > 0) {
            categories = categories.concat(
              this.$config.albumCategories().map((cat) => {
                return { value: cat, text: cat };
              })
            );
          }

          this.categories = categories;

          break;
        case "deleted":
          this.dirty = true;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];
            const index = this.results.findIndex((m) => m.UID === uid);

            if (index >= 0) {
              this.results.splice(index, 1);
            }

            this.removeSelection(uid);
          }

          break;
        case "created":
          this.dirty = true;

          for (let i = 0; i < data.entities.length; i++) {
            const values = data.entities[i];
            const index = this.results.findIndex((m) => m.UID === values.UID);

            if (index === -1 && this.staticFilter.type === values.Type) {
              this.results.unshift(new Album(values));
            }
          }
          break;
        default:
          console.warn("unexpected event type", ev);
      }
    },
  },
};
</script>
