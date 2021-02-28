package spark.sql.dialect.logical

import spark.sql.dialect.executor.DialectCommandOperator

trait DialectRunnableCommand extends org.apache.spark.sql.catalyst.plans.logical.Command {
  def run(operator: DialectCommandOperator): String
}
