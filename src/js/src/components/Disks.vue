<template>
    <div class="content">
      
      <template>
        <hr>
        <b-field grouped position="is-right">
          <p v-if="roleAllowed('disks', 'post')" class="control">
            <b-tooltip label="Upload a disk" type="is-light is-left">
              <button class="button is-light" @click="isCreateActive = true">
                <b-icon icon="upload"></b-icon>
              </button>
            </b-tooltip>
          </p>
        </b-field>
        <div>
          <b-table
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
            <b-table-column field="path" label="Path" centered v-slot="props">
              <b-tooltip :label="props.row.fullPath" type="is-dark">
                <b-icon icon="info-circle" size="is-small" />
              </b-tooltip>
            </b-table-column>
            <b-table-column field="backingImages" label="Backing Image" v-slot="props">
              <b-tooltip v-if="props.row.backingImages" type="is-dark" size="is-large" multilined>
                <template v-slot:content>
                    {{ props.row.name }}
                    <div v-for="i in props.row.backingImages">
                      &darr;<br>
                      {{  i  }}
                    </div>
                </template>
                {{ props.row.backingImages[0] }}
              </b-tooltip>
            </b-table-column>
            <b-table-column field="backingFor" label="Parent" centered v-slot="props">
              <b-tooltip :label="props.row.backingFor" type="is-dark">
                <b-icon v-if="props.row.backingFor" icon="chevron-circle-up" size="is-small" />
              </b-tooltip>
            </b-table-column>
            <b-table-column field="inUse" label="In Use" centered v-slot="props">
                <b-icon v-if="props.row.inUse" icon="play-circle" size="is-small" />
            </b-table-column>
            <b-table-column field="size" label="Size" v-slot="props">
              {{ props.row.size | fileSize }}
            </b-table-column>
            <b-table-column label="" width="20" v-slot="props">
              <b-tooltip label="Actions" type="is-dark">
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

              </b-tooltip>
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
          disks: []
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
  