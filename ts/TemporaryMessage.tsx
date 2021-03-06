import React, { Component } from 'react';
import * as action from './action';

const hideTimeOut = 5 * 1000; // 5 secs

interface State {
  message?: string;
}

export default class TemporaryMessage extends Component<any, State> {
  currTimerID: number;

  constructor(props?: any, context?: any) {
    super(props, context);

    this.showMessage = this.showMessage.bind(this);

    this.currTimerID = null;

    this.state = {
      message: null,
    };
  }

  componentDidMount() {
    action.onShowTemporaryMessage(this.showMessage, this);
  }

  componentWillUnmount() {
    action.offAllForOwner(this);
  }

  showMessage(msg: string, delay?: number) {
    this.setState({
      message: msg,
    });
    clearTimeout(this.currTimerID);
    if (delay) {
      this.currTimerID = setTimeout(() => this.showMessage(msg, delay));
      return;
    }
    this.currTimerID = window.setTimeout(() => {
      this.setState({
        message: null,
      });
    }, hideTimeOut);
  }

  render() {
    if (!this.state.message) {
      return <div className="hidden" />;
    }
    const html = {
      __html: this.state.message,
    };
    return <div className="temporary-message" dangerouslySetInnerHTML={html} />;
  }
}
