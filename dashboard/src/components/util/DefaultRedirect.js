import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

import SigninRedirect from "./SigninRedirect";
import LastEnvironmentRedirect from "./LastEnvironmentRedirect";

const query = gql`
  query DefaultRedirectQuery {
    auth @client {
      accessToken
      invalid
    }
  }
`;

class DefaultRedirect extends React.PureComponent {
  static propTypes = {
    data: PropTypes.object.isRequired,
  };

  render() {
    if (!this.props.data.auth.accessToken || this.props.data.auth.invalid) {
      return <SigninRedirect />;
    }
    return <LastEnvironmentRedirect />;
  }
}

export default graphql(query)(DefaultRedirect);
