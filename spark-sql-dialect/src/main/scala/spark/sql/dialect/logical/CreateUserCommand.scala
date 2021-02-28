package spark.sql.dialect.logical

import spark.sql.dialect.executor.DialectCommandOperator

case class CreateUserCommand(name: String, password: String,
                             sets: java.util.Map[String,String])
  extends DialectRunnableCommand {
  override def run(operator: DialectCommandOperator): String = {
    try {
      operator.createUser(name, password, sets)
    } catch {
      case e: Exception => {
        e.printStackTrace()
        e.getMessage
      }
    }
  }
}
