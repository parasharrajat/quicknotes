var Note = require('./Note.jsx');
var NewNoteSmall = require('./NewNoteSmall.jsx');

var NotesList = React.createClass({

  render: function () {
    var self = this;
    return (
      <div className="notes-list">
        <NewNoteSmall createNewTextNoteCb={this.props.createNewTextNoteCb}/>
        {this.props.notes.map(function(note) {
          return <Note
            compact={self.props.compact}
            note={note}
            key={note.IDStr}
            delUndelNoteCb={self.props.delUndelNoteCb}
            makeNotePublicPrivateCb={self.props.makeNotePublicPrivateCb}
          />;
        })}
      </div>
    );
  }
});

module.exports = NotesList;
