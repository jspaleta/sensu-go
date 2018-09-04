import React from "react";
import PropTypes from "prop-types";

import { Consumer } from "/components/relocation/Relocation";

class ToastProvider extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return (
      <Consumer>
        {({ createChild }) =>
          this.props.children({
            // TODO: Identify created element as a toast.
            addToast: render => createChild({ render }),
          })
        }
      </Consumer>
    );
  }
}

export default ToastProvider;
