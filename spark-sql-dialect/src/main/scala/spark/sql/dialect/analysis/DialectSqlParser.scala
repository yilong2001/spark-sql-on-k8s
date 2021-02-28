package spark.sql.dialect.analysis

import org.antlr.v4.runtime.misc.{Interval, ParseCancellationException}
import org.antlr.v4.runtime.tree.TerminalNodeImpl
import org.antlr.v4.runtime._
import org.antlr.v4.runtime.atn.PredictionMode
import org.apache.log4j.Logger
import org.apache.spark.sql.catalyst.parser.ParseErrorListener
import org.apache.spark.sql.catalyst.plans.logical.LogicalPlan
import org.apache.spark.sql.execution.SparkSqlParser
import org.apache.spark.sql.internal.SQLConf
import spark.sql.dialect.parser._

/**
  * Created by yilong on 2021/1/28.
  */
class DialectSqlParser(val conf: SQLConf) extends org.apache.spark.sql.catalyst.parser.ParserInterface {
  val logger = Logger.getLogger(classOf[DialectSqlParser])
  val baseSqlParser = new SparkSqlParser(conf)

  def parse[T](command: String)(toResult: SqlBaseParser => T): T = {
    logger.info(s"Parsing command: $command")

    val lexer = new SqlBaseLexer(new UpperCaseCharStream(CharStreams.fromString(command)))
    lexer.removeErrorListeners()
    lexer.addErrorListener(ParseErrorListener)

    val tokenStream = new CommonTokenStream(lexer)
    val parser = new SqlBaseParser(tokenStream)
    logger.info(parser.toString)
    parser.addParseListener(PostProcessor)
    parser.removeErrorListeners()
    parser.addErrorListener(ParseErrorListener)

    try {
      try {
        // first, try parsing with potentially faster SLL mode
        parser.getInterpreter.setPredictionMode(PredictionMode.SLL)
        toResult(parser)
      } catch {
        case e: ParseCancellationException =>
          // if we fail, parse with LL mode
          tokenStream.seek(0) // rewind input stream
          parser.reset()

          // Try Again.
          parser.getInterpreter.setPredictionMode(PredictionMode.LL)
          toResult(parser)
      }
    } catch {
      case e =>
        throw e
    }
  }

  def parseDataType(sqlText: String): org.apache.spark.sql.types.DataType = {
    baseSqlParser.parseDataType(sqlText)
  }
  def parseExpression(sqlText: String): org.apache.spark.sql.catalyst.expressions.Expression = {
    baseSqlParser.parseExpression(sqlText)
  }
  def parseFunctionIdentifier(sqlText: String): org.apache.spark.sql.catalyst.FunctionIdentifier = {
    baseSqlParser.parseFunctionIdentifier(sqlText)
  }
  def parseMultipartIdentifier(sqlText: String): Seq[String] = {
    baseSqlParser.parseMultipartIdentifier(sqlText)
  }
  def parsePlan(sqlText: String): org.apache.spark.sql.catalyst.plans.logical.LogicalPlan = {
    baseSqlParser.parsePlan(sqlText)
  }
  def parseRawDataType(sqlText: String): org.apache.spark.sql.types.DataType = {
    baseSqlParser.parseRawDataType(sqlText)
  }
  def parseTableIdentifier(sqlText: String): org.apache.spark.sql.catalyst.TableIdentifier = {
    baseSqlParser.parseTableIdentifier(sqlText)
  }
  def parseTableSchema(sqlText: String): org.apache.spark.sql.types.StructType = {
    baseSqlParser.parseTableSchema(sqlText)
  }
}


class UpperCaseCharStream(wrapped: CodePointCharStream) extends CharStream {
  override def consume(): Unit = wrapped.consume
  override def getSourceName(): String = wrapped.getSourceName
  override def index(): Int = wrapped.index
  override def mark(): Int = wrapped.mark
  override def release(marker: Int): Unit = wrapped.release(marker)
  override def seek(where: Int): Unit = wrapped.seek(where)
  override def size(): Int = wrapped.size

  override def getText(interval: Interval): String = {
    // ANTLR 4.7's CodePointCharStream implementations have bugs when
    // getText() is called with an empty stream, or intervals where
    // the start > end. See
    // https://github.com/antlr/antlr4/commit/ac9f7530 for one fix
    // that is not yet in a released ANTLR artifact.
    if (size() > 0 && (interval.b - interval.a >= 0)) {
      wrapped.getText(interval)
    } else {
      ""
    }
  }

  override def LA(i: Int): Int = {
    val la = wrapped.LA(i)
    if (la == 0 || la == IntStream.EOF) la
    else Character.toUpperCase(la)
  }
}

case object PostProcessor extends SqlBaseBaseListener {
  /** Remove the back ticks from an Identifier. */
  override def exitQuotedIdentifier(ctx: SqlBaseParser.QuotedIdentifierContext): Unit = {
    replaceTokenByIdentifier(ctx, 1) { token =>
      // Remove the double back ticks in the string.
      token.setText(token.getText.replace("``", "`"))
      token
    }
  }

  /** Treat non-reserved keywords as Identifiers. */
  override def exitNonReserved(ctx: SqlBaseParser.NonReservedContext): Unit = {
    replaceTokenByIdentifier(ctx, 0)(identity)
  }

  private def replaceTokenByIdentifier(ctx: ParserRuleContext,
                                       stripMargins: Int)(
                                       f: CommonToken => CommonToken = identity): Unit = {
    val parent = ctx.getParent
    parent.removeLastChild()
    val token = ctx.getChild(0).getPayload.asInstanceOf[Token]
    val newToken = new CommonToken(
      new org.antlr.v4.runtime.misc.Pair(token.getTokenSource, token.getInputStream),
      SqlBaseParser.IDENTIFIER,
      token.getChannel,
      token.getStartIndex + stripMargins,
      token.getStopIndex - stripMargins)
    parent.addChild(new TerminalNodeImpl(f(newToken)))
  }
}

object DialectSqlParser {
  def main(args: Array[String]): Unit = {
    val parser = new DialectSqlParser(new SQLConf)
    val astBuilder = new DialectSqlAstBuilder(null) //sparkSession.sqlContext.conf

    parser.parse("create user user1 identified by '123'") {
      parserHandler => {
        try {
          val stat = parserHandler.singleStatement()
          val res = astBuilder.visitSingleStatement(stat)
          //astBuilder.visitSingleStatement(parser.singleStatement()) match {
          res match {
            case plan: LogicalPlan => plan
            case _ => {
              throw new IllegalArgumentException()
            }
          }
        } catch {
          case e: Exception => {
            e.printStackTrace()
          }
        }
      }
    }

  }
}