import React from 'react';
import ReactDOM from 'react-dom';
import Top from './Top.jsx';
import ImportSimpleNote from './ImportSimpleNote.jsx';
import Editor from './Editor.jsx';

const AppNote = (props) => {
  console.log('appNoteStart: gLoggedInUserHandle: ', gLoggedInUserHandle);
  const isLoggedIn = gLoggedInUserHandle !== '';
  return (
    <div>
      <Top isLoggedIn={ isLoggedIn } loggedInUserHandle={ gLoggedInUserHandle } notesUserHandle="" />
      <ImportSimpleNote />
      <Editor />
    </div>
    );
};

function appNoteStart() {
  const el = document.getElementById('note-top');
  ReactDOM.render(<AppNote />, el);
}

window.appNoteStart = appNoteStart;
