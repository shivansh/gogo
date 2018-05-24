	.data
twoSpc:		.asciiz "  "
threeSpc:	.asciiz "   "
newline:		.asciiz "\n"
str:		.asciiz "Enter number of rows: "
rows:		.word	0
coef:		.word	0
i:		.word	0
space:		.word	0
temp:		.word	0
k:		.word	0

	.text
	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	li	$7, 1		# coef -> $7
	li	$15, 0		# i -> $15
	# Store variables back into memory
	sw	$3, rows
	sw	$7, coef
	sw	$15, i

outerFor:
	lw	$3, i
	lw	$7, rows
	bge	$3, $7, exit	# exit -> $0
	li	$30, 1		# space -> $30
	# Store variables back into memory
	sw	$3, i
	sw	$7, rows
	sw	$30, space

spcFor:
	lw	$3, rows
	lw	$7, i
	sub	$30, $3, $7	# temp -> $30
	li	$15, 0		# k -> $15
	lw	$16, space
	# Store variables back into memory
	sw	$3, rows
	sw	$7, i
	sw	$15, k
	sw	$16, space
	sw	$30, temp
	bgt	$16, $30, innerFor	# innerFor -> $0

	li	$v0, 4
	la	$a0, twoSpc
	syscall
	lw	$3, space
	addi	$3, $3, 1	# space -> $3
	# Store variables back into memory
	sw	$3, space
	j	spcFor

	li	$3, 0		# k -> $3
	# Store variables back into memory
	sw	$3, k

innerFor:
	lw	$3, k
	lw	$7, i
	bgt	$3, $7, endLine	# endLine -> $0
	# Store variables back into memory
	sw	$3, k
	sw	$7, i
	beq	$3, 0, labelIf	# labelIf -> $0

	lw	$3, i
	# Store variables back into memory
	sw	$3, i
	beq	$3, 0, labelIf	# labelIf -> $0

	lw	$3, i
	lw	$7, k
	sub	$30, $3, $7	# temp -> $30
	addi	$30, $30, 1	# temp -> $30
	lw	$15, coef
	mul	$15, $15, $30	# coef -> $15
	div	$15, $15, $7	# coef -> $15
	# Store variables back into memory
	sw	$3, i
	sw	$7, k
	sw	$15, coef
	sw	$30, temp
	j	labelCoef

labelIf:
	li	$3, 1		# coef -> $3
	# Store variables back into memory
	sw	$3, coef

labelCoef:
	li	$v0, 4
	la	$a0, threeSpc
	syscall
	li	$v0, 1
	lw	$3, coef
	move	$a0, $3
	syscall
	lw	$7, k
	addi	$7, $7, 1	# k -> $7
	# Store variables back into memory
	sw	$3, coef
	sw	$7, k
	j	innerFor

endLine:
	li	$v0, 4
	la	$a0, newline
	syscall
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	j	outerFor

exit:
	li	$v0, 10
	syscall
	.end main
