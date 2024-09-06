<template>
    <div class="content">    
      <b-modal :active.sync="detailsModal.active" :on-cancel="() => detailsModal.active = false" has-modal-card>
        <div class="modal-card">
          <header class="modal-card-head">
            <p class="modal-card-title">{{ detailsModal.disk.name }}</p>
          </header>
          <section class="modal-card-body">
            <p class="title is-5">Details</p>
            
          </section>
        </div>
        <!-- <b-tooltip label="Actions" type="is-dark">
                <b-dropdown aria-role="list">
                  <template #trigger>
                    <button class="button is-light is-small action">
                      <b-icon icon="ellipsis-v" custom-class="fa-fw fa-2xs"></b-icon>
                    </button>
                  </template>


                  <b-dropdown-item aria-role="listitem">Snapshot</b-dropdown-item>
                  <b-dropdown-item aria-role="listitem">Commit</b-dropdown-item>
                  <b-dropdown-item aria-role="listitem">Rebase</b-dropdown-item>
                  <b-dropdown-item aria-role="listitem">Clone</b-dropdown-item>
                  <hr class="dropdown-divider">
                  <b-dropdown-item aria-role="listitem">Download</b-dropdown-item>
                  <b-dropdown-item aria-role="listitem">Rename</b-dropdown-item>
                  <b-dropdown-item aria-role="listitem">Delete</b-dropdown-item>
                </b-dropdown>

              </b-tooltip> -->
      </b-modal>
      <template>
        <hr>
        <b-field grouped position="is-right">
          <b-tooltip label="Refresh List" type="is-light is-left">
            <button class="button is-light" @click="updateDisks">
              <b-icon icon="refresh"></b-icon>
            </button>
          </b-tooltip>
          <b-tooltip v-if="roleAllowed('disks', 'post')" label="Upload a disk" type="is-light is-left">
            <button class="button is-light" @click="isCreateActive = true" style="margin-left: 8px;">
              <b-icon icon="upload"></b-icon>
            </button>
          </b-tooltip>
        </b-field>
        <div>
          <b-table
            hoverable
            @click="rowClick"
            :row-class="(r, i) => 'is-clickable'"`
            :data="disks"
            :paginated="table.isPaginated"
            :per-page="table.perPage"
            :current-page.sync="table.currentPage"
            :pagination-simple="table.isPaginationSimple"
            :pagination-size="table.paginationSize"
            :default-sort-direction="table.defaultSortDirection"
            default-sort="name">
            <b-table-column field="name" label="Name" v-slot="props">
              {{ props.row.name }}
            </b-table-column>
            <b-table-column field="kind" label="Kind" v-slot="props">
              {{ props.row.kind }}
            </b-table-column>
            <b-table-column field="inUse" label="In Use" centered v-slot="props">
                <b-icon v-if="props.row.inUse" icon="play-circle" size="is-small" />
            </b-table-column>
            <b-table-column field="size" label="Size" v-slot="props">
              {{ props.row.size | fileSize }}
            </b-table-column>
          </b-table>
          <br>
          <b-field v-if="paginationNeeded" grouped position="is-right">
            <div class="control is-flex">
              <b-switch v-model="table.isPaginated" size="is-small" type="is-light" @input="changePaginate()">Paginate</b-switch>
            </div>
          </b-field>
        </div>
      </template>
      <!-- <b-loading :is-full-page="true" :active.sync="isWaiting" :can-cancel="false"></b-loading> -->
    </div>
  </template>
  
  <script>
    export default {
      
      beforeDestroy () {
      },
      
      async created () {
        this.updateDisks();
      },
  
      computed: {
        paginationNeeded () {
          var user = localStorage.getItem( 'user' );
  
          if ( localStorage.getItem( user + '.lastPaginate' ) ) {
            this.table.isPaginated = localStorage.getItem( user + '.lastPaginate' )  == 'true';
          }
  
          if ( this.disks.length <= 10 ) {
            this.table.isPaginated = false;
            return false;
          } else {
            return true;
          }
        }
      },
  
      methods: {
        updateDisks () {
          this.$http.get( 'disks' ).then(
            response => {
              console.log(response)
              response.json().then(
                state => {
                  console.log(state)
                  if ( state.disks.length == 0 ) {
                    this.isWaiting = true;
                  } else {
                    for ( let i = 0; i < state.disks.length; i++ ) {
                      this.disks.push( state.disks[ i ]);
                    }

                    this.disks.sort()
                    this.isWaiting = false;
                  }
                }
              );
            }, err => {
              this.errorNotification(err);
            }
          );
        },
        rowClick(row) {
          console.log(row)
          this.detailsModal.disk = row
          this.detailsModal.active = true
        }
      },
  
      data() {
        return {
          table: {
            striped: true,
            isPaginated: false,
            isPaginationSimple: true,
            paginationSize: 'is-small',
            defaultSortDirection: 'asc',
            currentPage: 1,
            perPage: 10
          },
          disks: [],
          detailsModal: {
            active: false,
            disk: {}
          },
        }
      }
    }
  </script>
  
  <style scoped>
    .b-tooltip:after {
      white-space: pre !important;
    }
    .dropdown-item {
      color: black !important;
    }
  </style>
  