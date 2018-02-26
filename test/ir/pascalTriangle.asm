	.data
twoSpc:	.asciiz "  "
threeSpc:	.asciiz "   "
newline:	.asciiz "\n"
str:	.asciiz "Enter number of rows: "
rows:	.word	0
coef:	.word	0
i:	.word	0
space:	.word	0
temp:	.word	0
k:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 1		# coef -> $t4
	sw $t1, rows		# spilled rows, freed $t1
	li $t1, 0		# i -> $t1
	# Store variables back into memory
	sw $t1, i
	sw $t4, coef

outerFor:

	lw $t1, i
	lw $t4, rows
	bge $t1, $t4, exit		# exit -> $t0
	li $t3, 1		# space -> $t3
	# Store variables back into memory
	sw $t1, i
	sw $t4, rows
	sw $t3, space

spcFor:
	lw $t1, rows
	lw $t4, i
	sub $t3, $t1, $t4	# temp -> $t3
	li $t2, 0		# k -> $t2
	# Store variables back into memory
	sw $t1, rows
	sw $t4, i
	sw $t3, temp
	sw $t2, k

	lw $t1, space
	lw $t4, temp
	bgt $t1, $t4, innerFor		# innerFor -> $t0
	li $v0, 4
	la $a0, twoSpc
	syscall
	addi $t1, $t1, 1		# space -> $t1
	# Store variables back into memory
	sw $t1, space
	sw $t4, temp

	j spcFor
	li $t1, 0		# k -> $t1
	# Store variables back into memory
	sw $t1, k

innerFor:

	lw $t1, k
	lw $t4, i
	bgt $t1, $t4, endLine		# endLine -> $t0
	# Store variables back into memory
	sw $t1, k
	sw $t4, i

	lw $t1, k
	beq $t1, 0, labelIf		# labelIf -> $t0
	# Store variables back into memory
	sw $t1, k

	lw $t1, i
	beq $t1, 0, labelIf		# labelIf -> $t0
	lw $t4, k
	sub $t3, $t1, $t4	# temp -> $t3
	addi $t3, $t3, 1		# temp -> $t3
	lw $t2, coef
	mul $t2, $t2, $t3	# coef -> $t2
	div $t2, $t2, $t4	# coef -> $t2
	# Store variables back into memory
	sw $t1, i
	sw $t4, k
	sw $t3, temp
	sw $t2, coef

	j labelCoef

labelIf:
	li $t1, 1		# coef -> $t1
	# Store variables back into memory
	sw $t1, coef

labelCoef:
	li $v0, 4
	la $a0, threeSpc
	syscall
	li $v0, 1
	lw $t1, coef
	move $a0, $t1
	syscall
	lw $t4, k
	addi $t4, $t4, 1		# k -> $t4
	# Store variables back into memory
	sw $t4, k
	sw $t1, coef

	j innerFor

endLine:
	li $v0, 4
	la $a0, newline
	syscall
	lw $t1, i
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i

	j outerFor

exit:
	li $v0, 10
	syscall
	.end main
