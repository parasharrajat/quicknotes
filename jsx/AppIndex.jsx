/* jshint -W097,-W117 */
'use strict';

var action = require('./action.js');
var Top = require('./Top.jsx');
var Settings = require('./Settings.jsx');
var u = require('./utils.js');

var RecentNotes = React.createClass({
  getInitialState: function() {
    return {
      notes: gRecentNotesInitial
    };
  },

  renderNotes: function(notes) {
    notes = u.arrNotNull(notes);
    return notes.map(function(note) {
      // see NoteSummary in db.go for note definition
      var userName = note.UserName;
      var title = note.Title;
      var k = note.IDStr;
      return <div key={k}>
        <span>{userName}</span>&nbsp;
        <span>{title}</span>
      </div>;
    });
  },

  render: function() {
    var notes = this.state.notes;
    return <div id="recentNotes">{this.renderNotes(notes)}</div>;
  }
});

var AppIndex = React.createClass({

  getInitialState: function() {
    return {
      showingSettings: false
    };
  },

  showSettings: function() {
    console.log("showSettings");
    this.setState({
      showingSettings: true
    });
  },

  hideSettings: function() {
    console.log("hideSettings");
    this.setState({
      showingSettings: false
    });
  },

  componentDidMount: function() {
    this.cidShowSettings = action.onShowSettings(this.showSettings);
    this.cidHideSettings = action.onHideSettings(this.hideSettings);
  },

  componentWillUnmount: function() {
    action.onShowSettings(this.cidShowSettings);
    action.onHideSettings(this.cidHideSettings);
  },

  renderSettings: function() {
    console.log("renderSettings: ", this.state.showingSettings);
    if (this.state.showingSettings) {
      return <Settings />;
    }
  },

  render: function() {
    console.log("AppIndex: gLoggedInUserHandle: ", gLoggedInUserHandle);
    var isLoggedIn = gLoggedInUserHandle !== "";
    return (
      <div>
        <Top isLoggedIn={isLoggedIn}
          loggedInUserHandle={gLoggedInUserHandle}
          notesUserHandle="" />
        {this.renderSettings()}
      </div>
    );
  }
});

function appIndexStart() {
  React.render(
    <AppIndex />,
    document.getElementById('root')
  );
  React.render(
    <RecentNotes />,
    document.getElementById('recent-notes-wrapper')
  );
}

window.appIndexStart = appIndexStart;
