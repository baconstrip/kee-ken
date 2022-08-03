owariTemplate = `
<div class="text-center" id="owari-board">
    <h3>Category: {{ category["Name"] }}</h3>
    <alertbox v-bind:message="message" v-if="message"></alertbox>
    <form v-if="!submitted">
        <div class="form-group">
            <label for="bid">Bid Amount</label> 
            <input type="number" id="bid-amount" placeholder="Enter Bid">
        </div>
        <button class="btn btn-success" type="button" @click="submit()">Submit Bid</button>
    </form>

    <form v-if="prompt && !ansSubmitted">
        <h2>Question: {{ prompt.Question }}</h2>
        <label for="answer">Answer</label>
        <input type="text" id="owari-answer" placeholder="Enter your answer here">
        <button class="btn btn-danger" type="button" @click="sendAnswer()">Submit Answer</button>
    </form>
    <ul>
        <li v-for="(n, a) in answers" v-if="answers">
            <em>{{ a }}: </em>{{ n }} (Bid: {{ bids[a] }})
        </li>
    </ul>
</div>
`

Vue.component('owari', {
    template: owariTemplate,
    props: ['category', 'host', 'prompt', 'answers', 'bids', 'money'],
    data: function () {
        return {
            'submitted': false,
            'ansSubmitted': false,
            'message': null,
        }
    },
    methods: {
        submit: function () {
            var amount = Number($('#bid-amount').val());
            if (amount < 0) {
                this.message = "Enter a positive value";
                return;
            }
            if (amount > this.money) {
                this.message = "Enter a value less than the amount you have";
                return;
            }
            this.$emit("submitBid", amount);
            this.submitted = true;
        },
        sendAnswer: function () {
            this.$emit("submitAnswer", $('#owari-answer').val());
            this.ansSubmitted = true;
        },
    },
});
