import React from "react";
import clsx from "clsx";
import styles from "./HomepageFeatures.module.css";
import Link from "@docusaurus/Link";

const FeatureList = [
  {
    title: "a9s CLI",
    Svg: require("../../static/img/Developers.svg").default,
    description: (
      <>
        A single Go binary to help with app development. Create local Kubernetes clusters and install a8s PostgreSQL with ease.
      </>
    ),
    button: {
      label: "a9s CLI Docs",
      link: "/docs/a9s-cli"
    }
  },
  {
    title: "Tutorial: Demo App + a8s PostgreSQL",
    Svg: require("../../static/img/Operator.svg").default,
    description: (
      <>
        Provision a local Kubernetes cluster using Kind or Minikube, install a PostgreSQL operator, deploy a demo application, load data into the database as well as perform backup and restore operations.
      </>
    ),
    button: {
      label: "Hands-On Tutorial",
      link: "/docs/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli/"
    }
  }
];

function Feature({ Svg, title, description, button }) {
  return (
    <div className={clsx("col col--6")}>
      <div className="text--center">
        <Svg className={styles.featureSvg} alt={title} />
      </div>
      <div className="text--center padding-horiz--md">
        <h2>{title}</h2>
        <p>{description}</p>
        <Link className="button button--secondary button--lg" to={button.link}>
          {button.label}
        </Link>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
