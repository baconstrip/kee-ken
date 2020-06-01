gameboardTemplate = `<div class="row"> 
  <div class="col col-lg-12" id="gameboard"> 
    <template v-if="board">
      <h2>Round: {{ board.Round }}</h2>
      <table class="table text-center">
        <thead>
          <tr>
            <th scope="col" v-for="category in board.Categories">
              {{ category.Name }}
            </th>
          </tr>
          <tr v-for="row in rows">
            <td v-for="question in row" @click="select" v-bind:qid="question.ID">
              <span v-if="!question.Played">{{ question.Value }}</span>
            </td>
          </tr>
        </thead>
      </table>
    </template>
    <div class="alert alert-primary" role="alert" v-else>
      Connecting to the server...
    </div>
  </div> 
</div>`

Vue.component('gameboard', {
    // board is the JSON representation of message.BoardOverview
    props: ['board', 'host'],
    data: function () {
        return {
            // rows are the rows of questions.
            rows: []
        }
    },
    template: gameboardTemplate,
    watch: {
        board: function(newVal, x) {
            console.log("updated");
            // Before we store the board, update it to sort questions by value.
            newVal.Categories.forEach(c => c.Questions.sort((a,b) => a.Value > b.Value));
            [...Array(5).keys()].forEach(i => 
                this.rows[i] = [...Array(5).keys()].map(j => newVal.Categories[j].Questions[i]));
        },
    },
    methods: {
        select: function (e) {
            var elem = e.srcElement;
            if (elem.tagName == "SPAN") {
                elem = elem.parentElement;
            }
            console.log(elem);
        }
    },
});
