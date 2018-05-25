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
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, rows	# spilled rows, freed $3
	li	$3, 1		# coef -> $3
	sw	$3, coef	# spilled coef, freed $3
	li	$3, 0		# i -> $3
	# Store variables back into memory
	sw	$3, i

outerFor:
	lw	$3, i
	lw	$5, rows
	bge	$3, $5, exit
	sw	$3, i		# spilled i, freed $3
	li	$3, 1		# space -> $3
	# Store variables back into memory
	sw	$3, space
	sw	$5, rows

spcFor:
	lw	$3, rows
	lw	$5, i
	sub	$6, $3, $5	# temp -> $6
	sw	$3, rows	# spilled rows, freed $3
	li	$3, 0		# k -> $3
	sw	$5, i		# spilled i, freed $5
	lw	$5, space
	# Store variables back into memory
	sw	$3, k
	sw	$5, space
	sw	$6, temp
	bgt	$5, $6, innerFor

	li	$2, 4
	la	$4, twoSpc
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
	lw	$5, i
	bgt	$3, $5, endLine
	# Store variables back into memory
	sw	$3, k
	sw	$5, i
	beq	$3, 0, labelIf

	lw	$3, i
	# Store variables back into memory
	sw	$3, i
	beq	$3, 0, labelIf

	lw	$3, i
	lw	$5, k
	sub	$6, $3, $5	# temp -> $6
	addi	$6, $6, 1	# temp -> $6
	sw	$3, i		# spilled i, freed $3
	lw	$3, coef
	mul	$3, $3, $6	# coef -> $3
	div	$3, $3, $5	# coef -> $3
	# Store variables back into memory
	sw	$3, coef
	sw	$5, k
	sw	$6, temp
	j	labelCoef

labelIf:
	li	$3, 1		# coef -> $3
	# Store variables back into memory
	sw	$3, coef

labelCoef:
	li	$2, 4
	la	$4, threeSpc
	syscall
	li	$2, 1
	lw	$3, coef
	move	$4, $3
	syscall
	sw	$3, coef	# spilled coef, freed $3
	lw	$3, k
	addi	$3, $3, 1	# k -> $3
	# Store variables back into memory
	sw	$3, k
	j	innerFor

endLine:
	li	$2, 4
	la	$4, newline
	syscall
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	j	outerFor

exit:
	li	$2, 10
	syscall
	.end main
